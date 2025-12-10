package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig  AppConfig  `env:"APPCONFIG"`
	DBConfig   DBConfig
	OIDCConfig OIDCConfig `env:"OIDC"`
}

type OIDCConfig struct {
	BaseURL                string         `env:"OIDC_BASE_URL" default:"http://localhost:3000"`
	FrontendURL            string         `env:"OIDC_FRONTEND_URL" default:"http://localhost:4321"`
	StateExpirationMinutes int            `env:"OIDC_STATE_EXPIRATION_MINUTES" default:"10"`
	Providers              []OIDCProvider `env:"-"` // Loaded separately from OIDC_PROVIDERS env var
}

// OIDCProvider represents an OIDC provider configuration
type OIDCProvider struct {
	Name         string `json:"name"`          // Display name (e.g., "Google", "Okta")
	Slug         string `json:"slug"`          // URL-safe identifier (e.g., "google", "okta")
	IssuerURL    string `json:"issuer_url"`    // OIDC issuer URL
	DiscoveryURL string `json:"discovery_url"` // Optional: different URL for discovery (Docker networking)
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scopes       string `json:"scopes"` // Space-separated scopes, defaults to "openid email profile"
}

type AppConfig struct {
	APPName            string `default:"pulse-api"`
	Port               int    `env:"PORT" default:"3000"`
	Version            string `default:"x.x.x" env:"VERSION"`
	Env                string `default:"development" env:"ENV"`
	JWTSecret          string `env:"JWT_SECRET" default:"dev-secret-change-in-production"`
	JWTExpirationHours int    `env:"JWT_EXPIRATION_HOURS" default:"24"`
}

type DBConfig struct {
	Host     string `default:"localhost" env:"DBHOST"`
	DataBase string `default:"pulse" env:"DBNAME"`
	User     string `default:"pulse" env:"DBUSERNAME"`
	Password string `required:"true" env:"DBPASSWORD" default:"mysecretpassword"`
	Port     uint   `default:"5432" env:"DBPORT"`
	SSLMode  string `default:"disable" env:"DBSSL"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	// Load OIDC providers from environment variable
	config.OIDCConfig.Providers = loadOIDCProviders()

	return config
}

// loadOIDCProviders loads OIDC provider configurations from the OIDC_PROVIDERS environment variable.
// The variable should be a JSON array of provider objects.
//
// Example:
//
//	OIDC_PROVIDERS='[{"name":"Google","slug":"google","issuer_url":"https://accounts.google.com","client_id":"...","client_secret":"..."}]'
//
// For multiple providers:
//
//	OIDC_PROVIDERS='[
//	  {"name":"Google","slug":"google","issuer_url":"https://accounts.google.com","client_id":"...","client_secret":"..."},
//	  {"name":"Okta","slug":"okta","issuer_url":"https://dev-123.okta.com","client_id":"...","client_secret":"..."}
//	]'
func loadOIDCProviders() []OIDCProvider {
	providersJSON := os.Getenv("OIDC_PROVIDERS")
	if providersJSON == "" {
		return nil
	}

	// Trim whitespace and handle multiline
	providersJSON = strings.TrimSpace(providersJSON)

	var providers []OIDCProvider
	if err := json.Unmarshal([]byte(providersJSON), &providers); err != nil {
		// Log error but don't panic - OIDC is optional
		return nil
	}

	// Set default scopes if not specified
	for i := range providers {
		if providers[i].Scopes == "" {
			providers[i].Scopes = "openid email profile"
		}
	}

	return providers
}
