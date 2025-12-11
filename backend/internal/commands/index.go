/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package commands

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/thatcatdev/kaimu/backend/internal/db"
	boardRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board"
	cardRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/card"
	orgRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization"
	orgMemberRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member"
	projectRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/project"
	userRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/logger"
	"github.com/thatcatdev/kaimu/backend/internal/services/search"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// stripHTML removes HTML tags from a string and normalizes whitespace
func stripHTML(s string) string {
	result := htmlTagRegex.ReplaceAllString(s, " ")
	result = strings.Join(strings.Fields(result), " ")
	return strings.TrimSpace(result)
}

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index all existing data to Typesense for search",
	Long:  `Bulk indexes all existing organizations, users, projects, boards, and cards to Typesense.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.LoadConfigOrPanic()

		logger.Logger(
			logger.WithServerName("kaimu-indexer"),
			logger.WithVersion("1.0.0"),
			logger.WithEnvironment(cfg.AppConfig.Env),
		)

		ctx := context.Background()
		log := logger.FromCtx(ctx)

		// Check if Typesense is configured
		if cfg.TypesenseConfig.Host == "" || cfg.TypesenseConfig.APIKey == "" {
			return fmt.Errorf("typesense is not configured. Set TYPESENSE_HOST and TYPESENSE_API_KEY environment variables")
		}

		// Initialize database
		database := db.NewDatabase(cfg.DBConfig)
		log.Info().Msg("Database connected")

		// Initialize Typesense client
		typesenseClient, err := search.NewTypesenseClient(cfg.TypesenseConfig)
		if err != nil {
			return fmt.Errorf("failed to create Typesense client: %w", err)
		}

		// Initialize repositories
		orgRepository := orgRepo.NewRepository(database.DB)
		orgMemberRepository := orgMemberRepo.NewRepository(database.DB)
		userRepository := userRepo.NewRepository(database.DB)
		projectRepository := projectRepo.NewRepository(database.DB)
		boardRepository := boardRepo.NewRepository(database.DB)
		cardRepository := cardRepo.NewRepository(database.DB)

		// Initialize search service
		searchService := search.NewService(typesenseClient, orgMemberRepository)

		// Initialize collections
		log.Info().Msg("Initializing Typesense collections...")
		if err := searchService.InitializeCollections(ctx); err != nil {
			return fmt.Errorf("failed to initialize collections: %w", err)
		}
		log.Info().Msg("Collections initialized")

		// Index organizations
		log.Info().Msg("Indexing organizations...")
		orgs, err := orgRepository.GetAll(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get organizations")
		} else {
			for _, org := range orgs {
				members, _ := orgMemberRepository.GetByOrgID(ctx, org.ID)
				memberIDs := make([]string, len(members))
				for i, m := range members {
					memberIDs[i] = m.UserID.String()
				}
				doc := &search.OrganizationDocument{
					ID:          org.ID.String(),
					Name:        org.Name,
					Slug:        org.Slug,
					Description: org.Description,
					OwnerID:     org.OwnerID.String(),
					MemberIDs:   memberIDs,
					CreatedAt:   org.CreatedAt.Unix(),
					UpdatedAt:   org.UpdatedAt.Unix(),
				}
				if err := searchService.IndexOrganization(ctx, doc); err != nil {
					log.Warn().Err(err).Str("org_id", org.ID.String()).Msg("Failed to index organization")
				}
			}
			log.Info().Int("count", len(orgs)).Msg("Organizations indexed")
		}

		// Index users
		log.Info().Msg("Indexing users...")
		users, err := userRepository.GetAll(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get users")
		} else {
			for _, user := range users {
				// Get user's organization memberships
				memberships, _ := orgMemberRepository.GetByUserID(ctx, user.ID)
				orgIDs := make([]string, len(memberships))
				for i, m := range memberships {
					orgIDs[i] = m.OrganizationID.String()
				}

				email := ""
				if user.Email != nil {
					email = *user.Email
				}
				displayName := ""
				if user.DisplayName != nil {
					displayName = *user.DisplayName
				}

				doc := &search.UserDocument{
					ID:              user.ID.String(),
					Username:        user.Username,
					Email:           email,
					DisplayName:     displayName,
					OrganizationIDs: orgIDs,
					CreatedAt:       user.CreatedAt.Unix(),
				}
				if err := searchService.IndexUser(ctx, doc); err != nil {
					log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("Failed to index user")
				}
			}
			log.Info().Int("count", len(users)).Msg("Users indexed")
		}

		// Index projects
		log.Info().Msg("Indexing projects...")
		projects, err := projectRepository.GetAll(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get projects")
		} else {
			// Build org maps for names and slugs
			orgNameMap := make(map[string]string)
			orgSlugMap := make(map[string]string)
			for _, org := range orgs {
				orgNameMap[org.ID.String()] = org.Name
				orgSlugMap[org.ID.String()] = org.Slug
			}

			for _, proj := range projects {
				doc := &search.ProjectDocument{
					ID:               proj.ID.String(),
					Name:             proj.Name,
					Key:              proj.Key,
					Description:      proj.Description,
					OrganizationID:   proj.OrganizationID.String(),
					OrganizationName: orgNameMap[proj.OrganizationID.String()],
					OrganizationSlug: orgSlugMap[proj.OrganizationID.String()],
					CreatedAt:        proj.CreatedAt.Unix(),
					UpdatedAt:        proj.UpdatedAt.Unix(),
				}
				if err := searchService.IndexProject(ctx, doc); err != nil {
					log.Warn().Err(err).Str("project_id", proj.ID.String()).Msg("Failed to index project")
				}
			}
			log.Info().Int("count", len(projects)).Msg("Projects indexed")
		}

		// Index boards
		log.Info().Msg("Indexing boards...")
		boards, err := boardRepository.GetAll(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get boards")
		} else {
			// Build project map
			projectMap := make(map[string]*projectRepo.Project)
			for _, proj := range projects {
				projectMap[proj.ID.String()] = proj
			}

			// Build org maps for names and slugs
			orgNameMap := make(map[string]string)
			orgSlugMap := make(map[string]string)
			for _, org := range orgs {
				orgNameMap[org.ID.String()] = org.Name
				orgSlugMap[org.ID.String()] = org.Slug
			}

			for _, board := range boards {
				proj := projectMap[board.ProjectID.String()]
				if proj == nil {
					continue
				}

				doc := &search.BoardDocument{
					ID:               board.ID.String(),
					Name:             board.Name,
					Description:      board.Description,
					IsDefault:        board.IsDefault,
					ProjectID:        proj.ID.String(),
					ProjectName:      proj.Name,
					ProjectKey:       proj.Key,
					OrganizationID:   proj.OrganizationID.String(),
					OrganizationName: orgNameMap[proj.OrganizationID.String()],
					OrganizationSlug: orgSlugMap[proj.OrganizationID.String()],
					CreatedAt:        board.CreatedAt.Unix(),
					UpdatedAt:        board.UpdatedAt.Unix(),
				}
				if err := searchService.IndexBoard(ctx, doc); err != nil {
					log.Warn().Err(err).Str("board_id", board.ID.String()).Msg("Failed to index board")
				}
			}
			log.Info().Int("count", len(boards)).Msg("Boards indexed")
		}

		// Index cards
		log.Info().Msg("Indexing cards...")
		cards, err := cardRepository.GetAll(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get cards")
		} else {
			// Build board map
			boardMap := make(map[string]*boardRepo.Board)
			for _, board := range boards {
				boardMap[board.ID.String()] = board
			}

			// Build project map
			projectMap := make(map[string]*projectRepo.Project)
			for _, proj := range projects {
				projectMap[proj.ID.String()] = proj
			}

			// Build org maps for names and slugs
			orgNameMap := make(map[string]string)
			orgSlugMap := make(map[string]string)
			for _, org := range orgs {
				orgNameMap[org.ID.String()] = org.Name
				orgSlugMap[org.ID.String()] = org.Slug
			}

			for _, card := range cards {
				board := boardMap[card.BoardID.String()]
				if board == nil {
					continue
				}
				proj := projectMap[board.ProjectID.String()]
				if proj == nil {
					continue
				}

				doc := &search.CardDocument{
					ID:               card.ID.String(),
					Title:            card.Title,
					Description:      stripHTML(card.Description),
					Priority:         string(card.Priority),
					BoardID:          board.ID.String(),
					BoardName:        board.Name,
					ProjectID:        proj.ID.String(),
					ProjectName:      proj.Name,
					ProjectKey:       proj.Key,
					OrganizationID:   proj.OrganizationID.String(),
					OrganizationName: orgNameMap[proj.OrganizationID.String()],
					OrganizationSlug: orgSlugMap[proj.OrganizationID.String()],
					Tags:             []string{}, // Could fetch tags here
					CreatedAt:        card.CreatedAt.Unix(),
					UpdatedAt:        card.UpdatedAt.Unix(),
				}

				if card.AssigneeID != nil {
					doc.AssigneeID = card.AssigneeID.String()
				}
				if card.DueDate != nil {
					doc.DueDate = card.DueDate.Unix()
				}

				if err := searchService.IndexCard(ctx, doc); err != nil {
					log.Warn().Err(err).Str("card_id", card.ID.String()).Msg("Failed to index card")
				}
			}
			log.Info().Int("count", len(cards)).Msg("Cards indexed")
		}

		log.Info().Msg("Indexing complete!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
