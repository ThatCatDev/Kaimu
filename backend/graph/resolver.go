package graph

import (
	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/thatcatdev/kaimu/backend/internal/resolvers"
	"github.com/thatcatdev/kaimu/backend/internal/services/audit"
	"github.com/thatcatdev/kaimu/backend/internal/services/auth"
	"github.com/thatcatdev/kaimu/backend/internal/services/board"
	"github.com/thatcatdev/kaimu/backend/internal/services/card"
	"github.com/thatcatdev/kaimu/backend/internal/services/email"
	"github.com/thatcatdev/kaimu/backend/internal/services/invitation"
	"github.com/thatcatdev/kaimu/backend/internal/services/metrics"
	"github.com/thatcatdev/kaimu/backend/internal/services/oidc"
	"github.com/thatcatdev/kaimu/backend/internal/services/organization"
	"github.com/thatcatdev/kaimu/backend/internal/services/project"
	"github.com/thatcatdev/kaimu/backend/internal/services/rbac"
	"github.com/thatcatdev/kaimu/backend/internal/services/search"
	"github.com/thatcatdev/kaimu/backend/internal/services/sprint"
	"github.com/thatcatdev/kaimu/backend/internal/services/tag"
	"github.com/thatcatdev/kaimu/backend/internal/services/user"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
// NOTE: Only services should be added here, not repositories.
// Repositories should be accessed through services.

type Resolver struct {
	Config                   config.Config
	AuthService              auth.Service
	AuditService             audit.Service
	OIDCService              oidc.Service
	OrganizationService      organization.Service
	ProjectService           project.Service
	BoardService             board.Service
	CardService              card.Service
	TagService               tag.Service
	RBACService              rbac.Service
	InvitationService        invitation.Service
	UserService              user.Service
	EmailVerificationService email.EmailVerificationService
	SearchService            search.Service
	SearchIndexer            *resolvers.SearchIndexer
	SprintService            sprint.Service
	MetricsService           metrics.Service
}
