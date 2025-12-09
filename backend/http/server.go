package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/thatcatdev/pulse-backend/config"
	"github.com/thatcatdev/pulse-backend/http/handlers"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/logger"
	"github.com/thatcatdev/pulse-backend/metrics"
	muxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

func SetupServer(cfg config.Config) *muxtrace.Router {

	router := muxtrace.NewRouter()

	// Add gzip compression middleware
	router.Use(middleware.GzipMiddleware())

	router.Handle("/ui/playground", playground.Handler("GraphQL playground", "/graphql")).Methods("GET")
	router.Handle("/graphql", handlers.BuildRootHandler(cfg)).Methods("POST")
	router.Handle("/healthcheck", handlers.HealthCheckHandler()).Methods("GET")
	router.Handle("/metrics", metrics.NewPrometheusInstance().Handler()).Methods("GET")

	return router
}

func SetupServerWithContext(ctx context.Context, cfg config.Config, deps *handlers.Dependencies) *muxtrace.Router {

	router := muxtrace.NewRouter(muxtrace.WithServiceName(cfg.AppConfig.APPName))

	// Add middleware to all routes - CORS must be first to handle preflight requests
	router.Use(middleware.CORSMiddleware([]string{"http://localhost:4321", "http://localhost:3000"}))
	router.Use(middleware.GzipMiddleware())
	router.Use(middleware.TracingMiddleware())
	router.Use(middleware.AuthMiddleware(deps.AuthService))

	router.Handle("/ui/playground", playground.Handler("GraphQL playground", "/graphql")).Methods("GET")
	router.Handle("/graphql", handlers.BuildRootHandlerWithContext(ctx, cfg, deps)).Methods("POST", "OPTIONS")
	router.Handle("/healthcheck", handlers.HealthCheckHandler()).Methods("GET")
	router.Handle("/metrics", metrics.NewPrometheusInstance().Handler()).Methods("GET")

	return router
}

func StartServer() error {
	cfg := config.LoadConfigOrPanic()
	router := SetupServer(cfg)

	log := logger.Get()
	log.Info().
		Int("port", cfg.AppConfig.Port).
		Str("playground_url", fmt.Sprintf("http://localhost:%d/", cfg.AppConfig.Port)).
		Msg("Starting GraphQL server")

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.Port), router)
}

func StartServerWithContext(ctx context.Context, deps *handlers.Dependencies) error {
	cfg := config.LoadConfigOrPanic()
	router := SetupServerWithContext(ctx, cfg, deps)

	log := logger.FromCtx(ctx)
	log.Info().
		Int("port", cfg.AppConfig.Port).
		Str("playground_url", fmt.Sprintf("http://localhost:%d/ui/playground", cfg.AppConfig.Port)).
		Msg("Starting GraphQL server")

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.AppConfig.Port), router)
}
