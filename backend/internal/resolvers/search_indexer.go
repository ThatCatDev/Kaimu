package resolvers

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	boardService "github.com/thatcatdev/kaimu/backend/internal/services/board"
	cardService "github.com/thatcatdev/kaimu/backend/internal/services/card"
	organizationService "github.com/thatcatdev/kaimu/backend/internal/services/organization"
	projectService "github.com/thatcatdev/kaimu/backend/internal/services/project"
	"github.com/thatcatdev/kaimu/backend/internal/services/search"
	userService "github.com/thatcatdev/kaimu/backend/internal/services/user"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// StripHTML removes HTML tags from a string and normalizes whitespace
func StripHTML(s string) string {
	// Remove HTML tags
	result := htmlTagRegex.ReplaceAllString(s, " ")
	// Normalize whitespace
	result = strings.Join(strings.Fields(result), " ")
	return strings.TrimSpace(result)
}

// SearchIndexer provides methods to index entities for search
// These methods are designed to be called asynchronously after CRUD operations
type SearchIndexer struct {
	searchSvc  search.Service
	orgSvc     organizationService.Service
	projectSvc projectService.Service
	boardSvc   boardService.Service
	cardSvc    cardService.Service
	userSvc    userService.Service
}

// NewSearchIndexer creates a new search indexer
func NewSearchIndexer(
	searchSvc search.Service,
	orgSvc organizationService.Service,
	projectSvc projectService.Service,
	boardSvc boardService.Service,
	cardSvc cardService.Service,
	userSvc userService.Service,
) *SearchIndexer {
	if searchSvc == nil {
		return nil
	}
	return &SearchIndexer{
		searchSvc:  searchSvc,
		orgSvc:     orgSvc,
		projectSvc: projectSvc,
		boardSvc:   boardSvc,
		cardSvc:    cardSvc,
		userSvc:    userSvc,
	}
}

// IndexOrganizationAsync indexes an organization asynchronously
func (si *SearchIndexer) IndexOrganizationAsync(ctx context.Context, orgID uuid.UUID, memberIDs []string) {
	if si == nil {
		return
	}
	go si.indexOrganization(context.Background(), orgID, memberIDs)
}

func (si *SearchIndexer) indexOrganization(ctx context.Context, orgID uuid.UUID, memberIDs []string) {
	org, err := si.orgSvc.GetOrganization(ctx, orgID)
	if err != nil {
		return
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

	_ = si.searchSvc.IndexOrganization(ctx, doc)
}

// DeleteOrganizationAsync deletes an organization from the index asynchronously
func (si *SearchIndexer) DeleteOrganizationAsync(ctx context.Context, orgID string) {
	if si == nil {
		return
	}
	go si.searchSvc.DeleteOrganization(context.Background(), orgID)
}

// IndexProjectAsync indexes a project asynchronously
func (si *SearchIndexer) IndexProjectAsync(ctx context.Context, projectID uuid.UUID) {
	if si == nil {
		return
	}
	go si.indexProject(context.Background(), projectID)
}

func (si *SearchIndexer) indexProject(ctx context.Context, projectID uuid.UUID) {
	proj, err := si.projectSvc.GetProject(ctx, projectID)
	if err != nil {
		return
	}

	// Get organization for name and slug
	org, err := si.orgSvc.GetOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return
	}

	doc := &search.ProjectDocument{
		ID:               proj.ID.String(),
		Name:             proj.Name,
		Key:              proj.Key,
		Description:      proj.Description,
		OrganizationID:   proj.OrganizationID.String(),
		OrganizationName: org.Name,
		OrganizationSlug: org.Slug,
		CreatedAt:        proj.CreatedAt.Unix(),
		UpdatedAt:        proj.UpdatedAt.Unix(),
	}

	_ = si.searchSvc.IndexProject(ctx, doc)
}

// DeleteProjectAsync deletes a project from the index asynchronously
func (si *SearchIndexer) DeleteProjectAsync(ctx context.Context, projectID string) {
	if si == nil {
		return
	}
	go si.searchSvc.DeleteProject(context.Background(), projectID)
}

// IndexBoardAsync indexes a board asynchronously
func (si *SearchIndexer) IndexBoardAsync(ctx context.Context, boardID uuid.UUID) {
	if si == nil {
		return
	}
	go si.indexBoard(context.Background(), boardID)
}

func (si *SearchIndexer) indexBoard(ctx context.Context, boardID uuid.UUID) {
	board, err := si.boardSvc.GetBoard(ctx, boardID)
	if err != nil {
		return
	}

	// Get project for org info
	proj, err := si.boardSvc.GetProject(ctx, boardID)
	if err != nil {
		return
	}

	// Get organization for name and slug
	org, err := si.orgSvc.GetOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return
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
		OrganizationName: org.Name,
		OrganizationSlug: org.Slug,
		CreatedAt:        board.CreatedAt.Unix(),
		UpdatedAt:        board.UpdatedAt.Unix(),
	}

	_ = si.searchSvc.IndexBoard(ctx, doc)
}

// DeleteBoardAsync deletes a board from the index asynchronously
func (si *SearchIndexer) DeleteBoardAsync(ctx context.Context, boardID string) {
	if si == nil {
		return
	}
	go si.searchSvc.DeleteBoard(context.Background(), boardID)
}

// IndexCardAsync indexes a card asynchronously
func (si *SearchIndexer) IndexCardAsync(ctx context.Context, cardID uuid.UUID) {
	if si == nil {
		return
	}
	go si.indexCard(context.Background(), cardID)
}

func (si *SearchIndexer) indexCard(ctx context.Context, cardID uuid.UUID) {
	card, err := si.cardSvc.GetCard(ctx, cardID)
	if err != nil {
		return
	}

	// Get board info
	board, err := si.cardSvc.GetBoardByCardID(ctx, cardID)
	if err != nil {
		return
	}

	// Get project info
	proj, err := si.boardSvc.GetProject(ctx, board.ID)
	if err != nil {
		return
	}

	// Get organization info
	org, err := si.orgSvc.GetOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return
	}

	// Get tags
	tags, _ := si.cardSvc.GetTagsForCard(ctx, cardID)
	tagNames := make([]string, len(tags))
	for i, t := range tags {
		tagNames[i] = t.Name
	}

	// Build document
	doc := &search.CardDocument{
		ID:               card.ID.String(),
		Title:            card.Title,
		Description:      StripHTML(card.Description),
		Priority:         string(card.Priority),
		BoardID:          board.ID.String(),
		BoardName:        board.Name,
		ProjectID:        proj.ID.String(),
		ProjectName:      proj.Name,
		ProjectKey:       proj.Key,
		OrganizationID:   proj.OrganizationID.String(),
		OrganizationName: org.Name,
		OrganizationSlug: org.Slug,
		Tags:             tagNames,
		CreatedAt:        card.CreatedAt.Unix(),
		UpdatedAt:        card.UpdatedAt.Unix(),
	}

	if card.AssigneeID != nil {
		doc.AssigneeID = card.AssigneeID.String()
		// Could fetch assignee name here if needed
	}

	if card.DueDate != nil {
		doc.DueDate = card.DueDate.Unix()
	}

	_ = si.searchSvc.IndexCard(ctx, doc)
}

// DeleteCardAsync deletes a card from the index asynchronously
func (si *SearchIndexer) DeleteCardAsync(ctx context.Context, cardID string) {
	if si == nil {
		return
	}
	go si.searchSvc.DeleteCard(context.Background(), cardID)
}

// IndexUserAsync indexes a user asynchronously
func (si *SearchIndexer) IndexUserAsync(ctx context.Context, userID uuid.UUID, orgIDs []string) {
	if si == nil {
		return
	}
	go si.indexUser(context.Background(), userID, orgIDs)
}

func (si *SearchIndexer) indexUser(ctx context.Context, userID uuid.UUID, orgIDs []string) {
	user, err := si.userSvc.GetByID(ctx, userID)
	if err != nil {
		return
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

	_ = si.searchSvc.IndexUser(ctx, doc)
}

// DeleteUserAsync deletes a user from the index asynchronously
func (si *SearchIndexer) DeleteUserAsync(ctx context.Context, userID string) {
	if si == nil {
		return
	}
	go si.searchSvc.DeleteUser(context.Background(), userID)
}

// Helper to get current timestamp
func nowUnix() int64 {
	return time.Now().Unix()
}
