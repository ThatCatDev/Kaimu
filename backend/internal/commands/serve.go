/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package commands

import (
	"context"

	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/thatcatdev/kaimu/backend/http"
	"github.com/thatcatdev/kaimu/backend/http/handlers"
	"github.com/thatcatdev/kaimu/backend/internal/logger"
	"github.com/thatcatdev/kaimu/backend/tracing"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the GraphQL API server",
	Long:  `Starts the Kaimu GraphQL API server with authentication support.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config to get environment
		cfg := config.LoadConfigOrPanic()

		// Initialize logger with environment
		logger.Logger(
			logger.WithServerName("kaimu-api"),
			logger.WithVersion("1.0.0"),
			logger.WithEnvironment(cfg.AppConfig.Env),
		)

		// Initialize tracing
		ctx := context.Background()
		tracedCtx, err := tracing.InitTracing(ctx)
		if err != nil {
			log := logger.FromCtx(ctx)
			log.Error().Err(err).Msg("Failed to initialize tracing")
			// Continue without tracing if initialization fails
			tracedCtx = ctx
		} else {
			defer func() {
				if err := tracing.Shutdown(context.Background()); err != nil {
					log := logger.FromCtx(tracedCtx)
					log.Error().Err(err).Msg("Error shutting down tracing")
				}
			}()
			log := logger.FromCtx(tracedCtx)
			log.Info().Msg("Tracing initialized successfully")
		}

		// Initialize all dependencies (database, repositories, services)
		deps := handlers.InitializeDependencies(cfg)
		log := logger.FromCtx(tracedCtx)
		log.Info().Msg("Dependencies initialized successfully")

		// Start the server with traced context
		return http.StartServerWithContext(tracedCtx, deps)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
