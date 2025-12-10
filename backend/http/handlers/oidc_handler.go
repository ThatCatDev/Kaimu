package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/logger"
	"github.com/thatcatdev/pulse-backend/internal/services/oidc"
)

type OIDCHandler struct {
	oidcService oidc.Service
	frontendURL string
	isSecure    bool
}

func NewOIDCHandler(oidcService oidc.Service, frontendURL string, isSecure bool) *OIDCHandler {
	return &OIDCHandler{
		oidcService: oidcService,
		frontendURL: frontendURL,
		isSecure:    isSecure,
	}
}

// ListProviders returns all enabled OIDC providers
// GET /auth/oidc/providers
func (h *OIDCHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	providers, err := h.oidcService.GetProviders(ctx)
	if err != nil {
		http.Error(w, "Failed to get providers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

// Authorize initiates the OIDC flow by redirecting to the provider
// GET /auth/oidc/{provider}/authorize
func (h *OIDCHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	providerSlug := vars["provider"]

	if providerSlug == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}

	// Get redirect URI from query params (optional, defaults to frontend)
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = h.frontendURL + "/dashboard"
	}

	authResponse, err := h.oidcService.GetAuthorizationURL(ctx, providerSlug, redirectURI)
	if err != nil {
		log := logger.FromCtx(ctx)
		log.Error().Err(err).Str("provider", providerSlug).Msg("Failed to get authorization URL")
		switch err {
		case oidc.ErrProviderNotFound:
			http.Error(w, "Provider not found", http.StatusNotFound)
		case oidc.ErrProviderDisabled:
			http.Error(w, "Provider is disabled", http.StatusForbidden)
		default:
			http.Error(w, "Failed to generate authorization URL", http.StatusInternalServerError)
		}
		return
	}

	// Redirect to OIDC provider
	http.Redirect(w, r, authResponse.AuthURL, http.StatusFound)
}

// Callback handles the OIDC callback from the provider
// GET /auth/oidc/{provider}/callback
func (h *OIDCHandler) Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	providerSlug := vars["provider"]

	if providerSlug == "" {
		h.redirectWithError(w, r, "Provider not specified")
		return
	}

	// Get code and state from query params
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		// Check for error response from provider
		errorParam := r.URL.Query().Get("error")
		errorDesc := r.URL.Query().Get("error_description")
		if errorParam != "" {
			h.redirectWithError(w, r, errorDesc)
			return
		}
		h.redirectWithError(w, r, "Missing code or state parameter")
		return
	}

	// Handle callback
	result, token, err := h.oidcService.HandleCallback(ctx, providerSlug, code, state)
	if err != nil {
		log := logger.FromCtx(ctx)
		log.Error().Err(err).Str("provider", providerSlug).Msg("OIDC callback failed")
		switch err {
		case oidc.ErrInvalidState, oidc.ErrStateExpired:
			h.redirectWithError(w, r, "Authentication session expired. Please try again.")
		case oidc.ErrTokenExchangeFailed:
			h.redirectWithError(w, r, "Failed to complete authentication. Please try again.")
		case oidc.ErrInvalidIDToken, oidc.ErrNonceMismatch:
			h.redirectWithError(w, r, "Invalid authentication response. Please try again.")
		default:
			h.redirectWithError(w, r, "Authentication failed. Please try again.")
		}
		return
	}

	// Set auth cookie
	middleware.SetAuthCookie(w, token, h.isSecure)

	// Determine redirect URL
	redirectURL := h.frontendURL + "/dashboard"

	// Add query param to indicate new user (for welcome flow)
	if result.IsNewUser {
		redirectURL += "?welcome=true"
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// redirectWithError redirects to the frontend login page with an error message
func (h *OIDCHandler) redirectWithError(w http.ResponseWriter, r *http.Request, message string) {
	// URL encode the error message
	redirectURL := h.frontendURL + "/login?error=" + message
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
