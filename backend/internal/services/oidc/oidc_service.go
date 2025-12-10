package oidc

//go:generate mockgen -source=oidc_service.go -destination=mocks/oidc_service_mock.go -package=mocks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/config"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/oidc_identity"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"golang.org/x/oauth2"
)

// ProviderInfo represents a simplified OIDC provider for API responses
type ProviderInfo struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// IdentityInfo represents a user's linked OIDC identity
type IdentityInfo struct {
	ProviderSlug string    `json:"provider_slug"`
	ProviderName string    `json:"provider_name"`
	Email        *string   `json:"email"`
	LinkedAt     time.Time `json:"linked_at"`
}

// AuthorizationResponse contains data needed to redirect to the OIDC provider
type AuthorizationResponse struct {
	AuthURL      string
	State        string
	CodeVerifier string
}

// CallbackResult contains the result of handling an OIDC callback
type CallbackResult struct {
	User             *user.User
	IsNewUser        bool
	LinkedToExisting bool
}

// Service interface for OIDC operations
type Service interface {
	// GetProviders returns all enabled OIDC providers
	GetProviders(ctx context.Context) ([]ProviderInfo, error)

	// GetAuthorizationURL generates an authorization URL for a provider
	GetAuthorizationURL(ctx context.Context, providerSlug, redirectURI string) (*AuthorizationResponse, error)

	// HandleCallback processes the OIDC callback, creates/links user, returns JWT token
	HandleCallback(ctx context.Context, providerSlug, code, state string) (*CallbackResult, string, error)

	// GetUserIdentities returns OIDC identities linked to a user
	GetUserIdentities(ctx context.Context, userID uuid.UUID) ([]IdentityInfo, error)

	// UnlinkIdentity removes an OIDC identity from a user
	UnlinkIdentity(ctx context.Context, userID uuid.UUID, providerSlug string) error
}

type service struct {
	providers    []config.OIDCProvider       // Providers from config
	providerMap  map[string]config.OIDCProvider // Lookup by slug
	identityRepo oidc_identity.Repository
	userRepo     user.Repository
	stateManager StateManager
	baseURL      string // Backend base URL for callbacks
	frontendURL  string // Frontend URL for redirects after auth
	jwtSecret    string
	jwtExpHours  int

	// Cache for OIDC providers (go-oidc Provider objects)
	oidcProviderCache map[string]*oidc.Provider
}

// NewService creates a new OIDC service
func NewService(
	providers []config.OIDCProvider,
	identityRepo oidc_identity.Repository,
	userRepo user.Repository,
	stateManager StateManager,
	baseURL, frontendURL, jwtSecret string,
	jwtExpHours int,
) Service {
	// Build provider map for fast lookup
	providerMap := make(map[string]config.OIDCProvider)
	for _, p := range providers {
		providerMap[p.Slug] = p
	}

	return &service{
		providers:         providers,
		providerMap:       providerMap,
		identityRepo:      identityRepo,
		userRepo:          userRepo,
		stateManager:      stateManager,
		baseURL:           baseURL,
		frontendURL:       frontendURL,
		jwtSecret:         jwtSecret,
		jwtExpHours:       jwtExpHours,
		oidcProviderCache: make(map[string]*oidc.Provider),
	}
}

func (s *service) GetProviders(ctx context.Context) ([]ProviderInfo, error) {
	result := make([]ProviderInfo, len(s.providers))
	for i, p := range s.providers {
		result[i] = ProviderInfo{
			Slug: p.Slug,
			Name: p.Name,
		}
	}
	return result, nil
}

func (s *service) getProviderBySlug(slug string) (*config.OIDCProvider, error) {
	provider, ok := s.providerMap[slug]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return &provider, nil
}

func (s *service) GetAuthorizationURL(ctx context.Context, providerSlug, redirectURI string) (*AuthorizationResponse, error) {
	// Get provider from config
	provider, err := s.getProviderBySlug(providerSlug)
	if err != nil {
		return nil, err
	}

	// Get OIDC provider metadata
	oidcProvider, err := s.getOIDCProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC provider metadata: %w", err)
	}

	// Create state with PKCE
	state, stateData, err := s.stateManager.CreateState(providerSlug, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create state: %w", err)
	}

	// Build OAuth2 config
	oauth2Config := s.buildOAuth2Config(provider, oidcProvider)

	// Generate code challenge for PKCE
	codeChallenge := GenerateCodeChallenge(stateData.CodeVerifier)

	// Build authorization URL with PKCE
	authURL := oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", stateData.Nonce),
	)

	return &AuthorizationResponse{
		AuthURL:      authURL,
		State:        state,
		CodeVerifier: stateData.CodeVerifier,
	}, nil
}

func (s *service) HandleCallback(ctx context.Context, providerSlug, code, state string) (*CallbackResult, string, error) {
	// Validate state and get PKCE data
	stateData, err := s.stateManager.GetState(state)
	if err != nil {
		return nil, "", err
	}

	// Delete state immediately (single use)
	s.stateManager.DeleteState(state)

	// Verify provider slug matches
	if stateData.ProviderSlug != providerSlug {
		return nil, "", ErrInvalidState
	}

	// Get provider from config
	provider, err := s.getProviderBySlug(providerSlug)
	if err != nil {
		return nil, "", err
	}

	// Get OIDC provider metadata
	oidcProvider, err := s.getOIDCProvider(ctx, provider)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get OIDC provider metadata: %w", err)
	}

	// Build OAuth2 config
	oauth2Config := s.buildOAuth2Config(provider, oidcProvider)

	// Exchange code for tokens with PKCE verifier
	token, err := oauth2Config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", stateData.CodeVerifier),
	)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrTokenExchangeFailed, err)
	}

	// Extract and verify ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, "", ErrInvalidIDToken
	}

	// Create verifier - use custom keyset if discovery URL differs from issuer URL
	var verifier *oidc.IDTokenVerifier
	if provider.DiscoveryURL != "" {
		// Create a custom RemoteKeySet that uses the discovery URL for JWKS
		// The JWKS URL is typically at {issuer}/keys for Dex
		jwksURL := provider.DiscoveryURL + "/keys"
		keySet := oidc.NewRemoteKeySet(ctx, jwksURL)
		verifier = oidc.NewVerifier(provider.IssuerURL, keySet, &oidc.Config{ClientID: provider.ClientID})
	} else {
		verifier = oidcProvider.Verifier(&oidc.Config{ClientID: provider.ClientID})
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrInvalidIDToken, err)
	}

	// Extract claims
	var claims struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Nonce         string `json:"nonce"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, "", fmt.Errorf("failed to parse claims: %w", err)
	}

	// Verify nonce
	if claims.Nonce != stateData.Nonce {
		return nil, "", ErrNonceMismatch
	}

	// Find or create user
	result, err := s.findOrCreateUser(ctx, provider, &claims)
	if err != nil {
		return nil, "", err
	}

	// Generate internal JWT token
	jwtToken, err := s.generateToken(result.User.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return result, jwtToken, nil
}

func (s *service) GetUserIdentities(ctx context.Context, userID uuid.UUID) ([]IdentityInfo, error) {
	identities, err := s.identityRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]IdentityInfo, 0, len(identities))
	for _, identity := range identities {
		// Find provider by issuer URL
		var providerName, providerSlug string
		for _, p := range s.providers {
			if p.IssuerURL == identity.Issuer {
				providerName = p.Name
				providerSlug = p.Slug
				break
			}
		}
		if providerSlug == "" {
			continue // Skip identities from unknown providers
		}

		result = append(result, IdentityInfo{
			ProviderSlug: providerSlug,
			ProviderName: providerName,
			Email:        identity.Email,
			LinkedAt:     identity.CreatedAt,
		})
	}

	return result, nil
}

func (s *service) UnlinkIdentity(ctx context.Context, userID uuid.UUID, providerSlug string) error {
	// Get provider to find issuer URL
	provider, err := s.getProviderBySlug(providerSlug)
	if err != nil {
		return err
	}

	return s.identityRepo.DeleteByUserIDAndIssuer(ctx, userID, provider.IssuerURL)
}

// Helper methods

func (s *service) getOIDCProvider(ctx context.Context, provider *config.OIDCProvider) (*oidc.Provider, error) {
	// Check cache first
	if oidcProv, ok := s.oidcProviderCache[provider.IssuerURL]; ok {
		return oidcProv, nil
	}

	var oidcProv *oidc.Provider
	var err error

	if provider.DiscoveryURL != "" {
		// Use InsecureIssuerURLContext when discovery URL is different from issuer URL
		// This is needed for Docker networking where:
		// - Browser needs to access http://localhost:5556/dex
		// - Backend needs to access http://dex:5556/dex
		insecureCtx := oidc.InsecureIssuerURLContext(ctx, provider.IssuerURL)
		oidcProv, err = oidc.NewProvider(insecureCtx, provider.DiscoveryURL)
	} else {
		// Standard case: discovery URL matches issuer URL
		oidcProv, err = oidc.NewProvider(ctx, provider.IssuerURL)
	}

	if err != nil {
		return nil, err
	}

	// Cache the provider
	s.oidcProviderCache[provider.IssuerURL] = oidcProv
	return oidcProv, nil
}

func (s *service) buildOAuth2Config(provider *config.OIDCProvider, oidcProvider *oidc.Provider) *oauth2.Config {
	scopes := strings.Split(provider.Scopes, " ")
	endpoint := oidcProvider.Endpoint()

	// If a discovery URL is set (for Docker networking), rewrite the endpoint URLs
	// to use the discovery URL's host instead of the issuer URL's host
	if provider.DiscoveryURL != "" {
		endpoint = s.rewriteEndpoint(endpoint, provider.IssuerURL, provider.DiscoveryURL)
	}

	return &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Endpoint:     endpoint,
		RedirectURL:  fmt.Sprintf("%s/auth/oidc/%s/callback", s.baseURL, provider.Slug),
		Scopes:       scopes,
	}
}

// rewriteEndpoint replaces the issuer host with the discovery host in TokenURL only
// AuthURL stays as-is since the browser needs to access it via localhost
// TokenURL is rewritten because the backend calls it from inside Docker
func (s *service) rewriteEndpoint(endpoint oauth2.Endpoint, issuerURL, discoveryURL string) oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  endpoint.AuthURL, // Keep original - browser needs localhost
		TokenURL: strings.Replace(endpoint.TokenURL, issuerURL, discoveryURL, 1),
	}
}

func (s *service) findOrCreateUser(ctx context.Context, provider *config.OIDCProvider, claims *struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Nonce         string `json:"nonce"`
}) (*CallbackResult, error) {
	// Check if identity already exists
	existingIdentity, err := s.identityRepo.GetByIssuerAndSubject(ctx, provider.IssuerURL, claims.Subject)
	if err == nil && existingIdentity != nil {
		// Identity exists, get the user
		u, err := s.userRepo.GetByID(ctx, existingIdentity.UserID)
		if err != nil {
			return nil, err
		}

		// Update user info if changed
		updated := false
		if claims.Email != "" && (u.Email == nil || *u.Email != claims.Email) {
			u.Email = &claims.Email
			updated = true
		}
		if claims.Name != "" && (u.DisplayName == nil || *u.DisplayName != claims.Name) {
			u.DisplayName = &claims.Name
			updated = true
		}
		if claims.Picture != "" && (u.AvatarURL == nil || *u.AvatarURL != claims.Picture) {
			u.AvatarURL = &claims.Picture
			updated = true
		}
		if updated {
			s.userRepo.Update(ctx, u)
		}

		return &CallbackResult{
			User:             u,
			IsNewUser:        false,
			LinkedToExisting: false,
		}, nil
	}

	// Identity doesn't exist - try to find user by email
	if claims.Email != "" && claims.EmailVerified {
		existingUser, err := s.userRepo.GetByEmail(ctx, claims.Email)
		if err == nil && existingUser != nil {
			// Link identity to existing user
			identity := &oidc_identity.OIDCIdentity{
				UserID:        existingUser.ID,
				Issuer:        provider.IssuerURL,
				Subject:       claims.Subject,
				Email:         &claims.Email,
				EmailVerified: claims.EmailVerified,
			}
			if err := s.identityRepo.Create(ctx, identity); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrIdentityLinkFailed, err)
			}

			// Update user info
			if claims.Name != "" && existingUser.DisplayName == nil {
				existingUser.DisplayName = &claims.Name
			}
			if claims.Picture != "" && existingUser.AvatarURL == nil {
				existingUser.AvatarURL = &claims.Picture
			}
			s.userRepo.Update(ctx, existingUser)

			return &CallbackResult{
				User:             existingUser,
				IsNewUser:        false,
				LinkedToExisting: true,
			}, nil
		}
	}

	// Create new user
	username := s.generateUsername(claims.Email, claims.Name, claims.Subject)
	newUser := &user.User{
		Username:    username,
		Email:       nilIfEmpty(claims.Email),
		DisplayName: nilIfEmpty(claims.Name),
		AvatarURL:   nilIfEmpty(claims.Picture),
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUserCreationFailed, err)
	}

	// Create OIDC identity
	identity := &oidc_identity.OIDCIdentity{
		UserID:        newUser.ID,
		Issuer:        provider.IssuerURL,
		Subject:       claims.Subject,
		Email:         nilIfEmpty(claims.Email),
		EmailVerified: claims.EmailVerified,
	}
	if err := s.identityRepo.Create(ctx, identity); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrIdentityLinkFailed, err)
	}

	return &CallbackResult{
		User:             newUser,
		IsNewUser:        true,
		LinkedToExisting: false,
	}, nil
}

func (s *service) generateUsername(email, name, subject string) string {
	// Try to use email prefix
	if email != "" {
		parts := strings.Split(email, "@")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0] + "_" + uuid.New().String()[:8]
		}
	}

	// Try to use name
	if name != "" {
		return strings.ReplaceAll(strings.ToLower(name), " ", "_") + "_" + uuid.New().String()[:8]
	}

	// Fall back to subject-based username
	return "user_" + uuid.New().String()[:8]
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Token generation (duplicated from auth service to avoid circular dependency)
func (s *service) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"iss":     "pulse",
		"exp":     time.Now().Add(time.Duration(s.jwtExpHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
