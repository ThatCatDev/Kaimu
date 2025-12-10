package oidc

import "errors"

var (
	ErrProviderNotFound     = errors.New("OIDC provider not found")
	ErrProviderDisabled     = errors.New("OIDC provider is disabled")
	ErrInvalidState         = errors.New("invalid or missing state parameter")
	ErrStateExpired         = errors.New("state parameter has expired")
	ErrTokenExchangeFailed  = errors.New("failed to exchange authorization code for tokens")
	ErrInvalidIDToken       = errors.New("invalid ID token")
	ErrNonceMismatch        = errors.New("ID token nonce does not match")
	ErrUserCreationFailed   = errors.New("failed to create user from OIDC identity")
	ErrIdentityLinkFailed   = errors.New("failed to link OIDC identity to user")
	ErrIdentityAlreadyLinked = errors.New("OIDC identity is already linked to another user")
)
