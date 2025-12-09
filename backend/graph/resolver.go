package graph

import (
	"github.com/thatcatdev/pulse-backend/config"
	"github.com/thatcatdev/pulse-backend/internal/services/auth"
	"github.com/thatcatdev/pulse-backend/internal/services/organization"
	"github.com/thatcatdev/pulse-backend/internal/services/project"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
// NOTE: Only services should be added here, not repositories.
// Repositories should be accessed through services.

type Resolver struct {
	Config              config.Config
	AuthService         auth.Service
	OrganizationService organization.Service
	ProjectService      project.Service
}
