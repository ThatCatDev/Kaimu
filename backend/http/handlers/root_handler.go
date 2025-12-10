package handlers

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/thatcatdev/pulse-backend/config"
	"github.com/thatcatdev/pulse-backend/graph"
	"github.com/thatcatdev/pulse-backend/graph/generated"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/db"
	boardRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/board"
	boardColumnRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/board_column"
	cardRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/card"
	cardTagRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/card_tag"
	oidcIdentityRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/oidc_identity"
	orgRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/organization"
	orgMemberRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/organization_member"
	projectRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	tagRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/tag"
	userRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"github.com/thatcatdev/pulse-backend/internal/directives"
	"github.com/thatcatdev/pulse-backend/internal/services/auth"
	"github.com/thatcatdev/pulse-backend/internal/services/board"
	"github.com/thatcatdev/pulse-backend/internal/services/card"
	"github.com/thatcatdev/pulse-backend/internal/services/oidc"
	"github.com/thatcatdev/pulse-backend/internal/services/organization"
	"github.com/thatcatdev/pulse-backend/internal/services/project"
	"github.com/thatcatdev/pulse-backend/internal/services/tag"
)

// Dependencies holds all initialized dependencies for the application
type Dependencies struct {
	AuthService         auth.Service
	OIDCService         oidc.Service
	OrganizationService organization.Service
	ProjectService      project.Service
	BoardService        board.Service
	CardService         card.Service
	TagService          tag.Service
	OIDCHandler         *OIDCHandler
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
		AuthService:         authService,
		OIDCService:         oidcService,
		OrganizationService: organizationService,
		ProjectService:      projectService,
		BoardService:        boardService,
		CardService:         cardService,
		TagService:          tagService,
		OIDCHandler:         oidcHandler,
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
		Config:              conf,
		AuthService:         deps.AuthService,
		OIDCService:         deps.OIDCService,
		OrganizationService: deps.OrganizationService,
		ProjectService:      deps.ProjectService,
		BoardService:        deps.BoardService,
		CardService:         deps.CardService,
		TagService:          deps.TagService,
	}

	cfg := generated.Config{Resolvers: resolvers, Directives: directives.GetDirectives()}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

	// Add GraphQL tracing extension
	srv.Use(&middleware.GraphQLTracingExtension{})

	return srv
}
