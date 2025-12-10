package oidc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"sync"
	"time"
)

// StateData holds PKCE and session data for an OAuth flow
type StateData struct {
	ProviderSlug string
	CodeVerifier string
	Nonce        string
	RedirectURI  string
	CreatedAt    time.Time
}

// StateManager handles PKCE state storage and validation
type StateManager interface {
	CreateState(providerSlug, redirectURI string) (state string, data *StateData, err error)
	GetState(state string) (*StateData, error)
	DeleteState(state string)
	Cleanup()
}

type inMemoryStateManager struct {
	states     map[string]*StateData
	mu         sync.RWMutex
	expiration time.Duration
}

// NewStateManager creates a new in-memory state manager
func NewStateManager(expirationMinutes int) StateManager {
	sm := &inMemoryStateManager{
		states:     make(map[string]*StateData),
		expiration: time.Duration(expirationMinutes) * time.Minute,
	}

	// Start cleanup goroutine
	go sm.cleanupLoop()

	return sm
}

func (sm *inMemoryStateManager) CreateState(providerSlug, redirectURI string) (string, *StateData, error) {
	// Generate cryptographically secure state parameter
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", nil, err
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	// Generate PKCE code verifier (43-128 characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return "", nil, err
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate nonce for ID token validation
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", nil, err
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)

	data := &StateData{
		ProviderSlug: providerSlug,
		CodeVerifier: codeVerifier,
		Nonce:        nonce,
		RedirectURI:  redirectURI,
		CreatedAt:    time.Now(),
	}

	sm.mu.Lock()
	sm.states[state] = data
	sm.mu.Unlock()

	return state, data, nil
}

func (sm *inMemoryStateManager) GetState(state string) (*StateData, error) {
	sm.mu.RLock()
	data, exists := sm.states[state]
	sm.mu.RUnlock()

	if !exists {
		return nil, ErrInvalidState
	}

	// Check expiration
	if time.Since(data.CreatedAt) > sm.expiration {
		sm.DeleteState(state)
		return nil, ErrStateExpired
	}

	return data, nil
}

func (sm *inMemoryStateManager) DeleteState(state string) {
	sm.mu.Lock()
	delete(sm.states, state)
	sm.mu.Unlock()
}

func (sm *inMemoryStateManager) Cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for state, data := range sm.states {
		if now.Sub(data.CreatedAt) > sm.expiration {
			delete(sm.states, state)
		}
	}
}

func (sm *inMemoryStateManager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		sm.Cleanup()
	}
}

// GenerateCodeChallenge generates a PKCE S256 code challenge from a code verifier
func GenerateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
