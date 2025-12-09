package graph

import (
	"github.com/thatcatdev/pulse-backend/config"
	"github.com/thatcatdev/pulse-backend/internal/services/auth"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Config      config.Config
	AuthService auth.Service
}
