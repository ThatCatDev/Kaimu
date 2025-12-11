package handlers

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/thatcatdev/kaimu/backend/graph"
	"github.com/thatcatdev/kaimu/backend/graph/generated"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	"github.com/thatcatdev/kaimu/backend/internal/db"
	boardRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board"
	boardColumnRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column"
	cardRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/card"
	cardTagRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/card_tag"
	emailVerificationTokenRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/email_verification_token"
	invitationRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/invitation"
	oidcIdentityRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/oidc_identity"
	orgRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization"
	orgMemberRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member"
	permissionRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/permission"
	projectRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/project"
	projectMemberRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/project_member"
	roleRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/role"
	rolePermissionRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/role_permission"
	tagRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/tag"
	userRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/directives"
	"github.com/thatcatdev/kaimu/backend/internal/services/auth"
	"github.com/thatcatdev/kaimu/backend/internal/services/board"
	"github.com/thatcatdev/kaimu/backend/internal/services/card"
	"github.com/thatcatdev/kaimu/backend/internal/services/email"
	"github.com/thatcatdev/kaimu/backend/internal/services/invitation"
	"github.com/thatcatdev/kaimu/backend/internal/services/mail"
	"github.com/thatcatdev/kaimu/backend/internal/services/mjml"
	"github.com/thatcatdev/kaimu/backend/internal/services/oidc"
	"github.com/thatcatdev/kaimu/backend/internal/services/organization"
	"github.com/thatcatdev/kaimu/backend/internal/services/project"
	"github.com/thatcatdev/kaimu/backend/internal/services/rbac"
	"github.com/thatcatdev/kaimu/backend/internal/services/tag"
	"github.com/thatcatdev/kaimu/backend/internal/services/user"
)

// Dependencies holds all initialized dependencies for the application
type Dependencies struct {
	AuthService              auth.Service
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
	OIDCHandler              *OIDCHandler
}

// InitializeDependencies creates all application dependencies
func InitializeDependencies(cfg config.Config) *Dependencies {
	// Initialize database
	database := db.NewDatabase(cfg.DBConfig)

	// Initialize repositories
	userRepository := userRepo.NewRepository(database.DB)
	orgRepository := orgRepo.NewRepository(database.DB)
	orgMemberRepository := orgMemberRepo.NewRepository(database.DB)
	projectRepository := projectRepo.NewRepository(database.DB)
	boardRepository := boardRepo.NewRepository(database.DB)
	boardColumnRepository := boardColumnRepo.NewRepository(database.DB)
	cardRepository := cardRepo.NewRepository(database.DB)
	tagRepository := tagRepo.NewRepository(database.DB)
	cardTagRepository := cardTagRepo.NewRepository(database.DB)
	oidcIdentityRepository := oidcIdentityRepo.NewRepository(database.DB)
	permissionRepository := permissionRepo.NewRepository(database.DB)
	roleRepository := roleRepo.NewRepository(database.DB)
	rolePermissionRepository := rolePermissionRepo.NewRepository(database.DB)
	projectMemberRepository := projectMemberRepo.NewRepository(database.DB)
	invitationRepository := invitationRepo.NewRepository(database.DB)

	// Initialize services
	authService := auth.NewService(
		userRepository,
		cfg.AppConfig.JWTSecret,
		cfg.AppConfig.JWTExpirationHours,
	)

	organizationService := organization.NewService(
		orgRepository,
		orgMemberRepository,
		userRepository,
	)

	projectService := project.NewService(
		projectRepository,
		orgRepository,
	)

	boardService := board.NewService(
		boardRepository,
		boardColumnRepository,
		projectRepository,
	)

	cardService := card.NewService(
		cardRepository,
		boardColumnRepository,
		boardRepository,
		tagRepository,
		cardTagRepository,
	)

	tagService := tag.NewService(
		tagRepository,
		projectRepository,
	)

	rbacService := rbac.NewService(
		permissionRepository,
		roleRepository,
		rolePermissionRepository,
		orgMemberRepository,
		projectMemberRepository,
		projectRepository,
		userRepository,
	)

	// Initialize email services first (needed by invitation service)
	emailVerificationTokenRepository := emailVerificationTokenRepo.NewEmailVerificationTokenRepository(database.DB)
	mjmlService := mjml.NewMJMLService()
	mailService := mail.NewMailService(cfg.EmailConfig, mjmlService)

	invitationService := invitation.NewService(
		invitationRepository,
		orgRepository,
		orgMemberRepository,
		userRepository,
		roleRepository,
		mailService,
		cfg.EmailConfig,
	)

	userService := user.NewService(userRepository)

	// Initialize email verification service (uses same mail service)
	emailVerificationService := email.NewEmailVerificationService(
		emailVerificationTokenRepository,
		userRepository,
		mailService,
		cfg.EmailConfig,
	)

	// Initialize OIDC service and handler
	stateManager := oidc.NewStateManager(cfg.OIDCConfig.StateExpirationMinutes)
	oidcService := oidc.NewService(
		cfg.OIDCConfig.Providers, // Providers from config (env var)
		oidcIdentityRepository,
		userRepository,
		stateManager,
		cfg.OIDCConfig.BaseURL,
		cfg.OIDCConfig.FrontendURL,
		cfg.AppConfig.JWTSecret,
		cfg.AppConfig.JWTExpirationHours,
	)

	isSecure := cfg.AppConfig.Env != "development"
	oidcHandler := NewOIDCHandler(oidcService, cfg.OIDCConfig.FrontendURL, isSecure)

	return &Dependencies{
		AuthService:              authService,
		OIDCService:              oidcService,
		OrganizationService:      organizationService,
		ProjectService:           projectService,
		BoardService:             boardService,
		CardService:              cardService,
		TagService:               tagService,
		RBACService:              rbacService,
		InvitationService:        invitationService,
		UserService:              userService,
		EmailVerificationService: emailVerificationService,
		OIDCHandler:              oidcHandler,
	}
}

func BuildRootHandler(conf config.Config) http.Handler {
	resolvers := &graph.Resolver{
		Config: conf,
	}

	cfg := generated.Config{Resolvers: resolvers, Directives: directives.GetDirectives()}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

	// Add GraphQL tracing extension
	srv.Use(&middleware.GraphQLTracingExtension{})

	return srv
}

func BuildRootHandlerWithContext(ctx context.Context, conf config.Config, deps *Dependencies) http.Handler {
	resolvers := &graph.Resolver{
		Config:                   conf,
		AuthService:              deps.AuthService,
		OIDCService:              deps.OIDCService,
		OrganizationService:      deps.OrganizationService,
		ProjectService:           deps.ProjectService,
		BoardService:             deps.BoardService,
		CardService:              deps.CardService,
		TagService:               deps.TagService,
		RBACService:              deps.RBACService,
		InvitationService:        deps.InvitationService,
		UserService:              deps.UserService,
		EmailVerificationService: deps.EmailVerificationService,
	}

	cfg := generated.Config{Resolvers: resolvers, Directives: directives.GetDirectives()}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

	// Add GraphQL tracing extension
	srv.Use(&middleware.GraphQLTracingExtension{})

	return srv
}
