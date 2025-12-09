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
	userRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"github.com/thatcatdev/pulse-backend/internal/directives"
	"github.com/thatcatdev/pulse-backend/internal/services/auth"
)

// Dependencies holds all initialized dependencies for the application
type Dependencies struct {
	AuthService auth.Service
}

// InitializeDependencies creates all application dependencies
func InitializeDependencies(cfg config.Config) *Dependencies {
	// Initialize database
	database := db.NewDatabase(cfg.DBConfig)

	// Initialize repositories
	userRepository := userRepo.NewRepository(database.DB)

	// Initialize services
	authService := auth.NewService(
		userRepository,
		cfg.AppConfig.JWTSecret,
		cfg.AppConfig.JWTExpirationHours,
	)

	return &Dependencies{
		AuthService: authService,
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
		Config:      conf,
		AuthService: deps.AuthService,
	}

	cfg := generated.Config{Resolvers: resolvers, Directives: directives.GetDirectives()}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

	// Add GraphQL tracing extension
	srv.Use(&middleware.GraphQLTracingExtension{})

	return srv
}
