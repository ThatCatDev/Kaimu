package graph

import (
	"github.com/thatcatdev/pulse-backend/config"
	"github.com/thatcatdev/pulse-backend/internal/services/auth"
	"github.com/thatcatdev/pulse-backend/internal/services/board"
	"github.com/thatcatdev/pulse-backend/internal/services/card"
	"github.com/thatcatdev/pulse-backend/internal/services/invitation"
	"github.com/thatcatdev/pulse-backend/internal/services/oidc"
	"github.com/thatcatdev/pulse-backend/internal/services/organization"
	"github.com/thatcatdev/pulse-backend/internal/services/project"
	"github.com/thatcatdev/pulse-backend/internal/services/rbac"
	"github.com/thatcatdev/pulse-backend/internal/services/tag"
	"github.com/thatcatdev/pulse-backend/internal/services/user"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
// NOTE: Only services should be added here, not repositories.
// Repositories should be accessed through services.

type Resolver struct {
	Config              config.Config
	AuthService         auth.Service
	OIDCService         oidc.Service
	OrganizationService organization.Service
	ProjectService      project.Service
	BoardService        board.Service
	CardService         card.Service
	TagService          tag.Service
	RBACService         rbac.Service
	InvitationService   invitation.Service
	UserService         user.Service
}
