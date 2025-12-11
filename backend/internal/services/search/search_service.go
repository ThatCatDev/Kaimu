package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member"
	"github.com/thatcatdev/kaimu/backend/tracing"
	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"
	"github.com/typesense/typesense-go/v2/typesense/api/pointer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Service defines the search service interface
type Service interface {
	// Search performs a multi-collection search with access control
	Search(ctx context.Context, userID uuid.UUID, query string, scope *SearchScope, limit int) (*SearchResults, error)

	// Indexing methods
	IndexOrganization(ctx context.Context, doc *OrganizationDocument) error
	IndexUser(ctx context.Context, doc *UserDocument) error
	IndexProject(ctx context.Context, doc *ProjectDocument) error
	IndexBoard(ctx context.Context, doc *BoardDocument) error
	IndexCard(ctx context.Context, doc *CardDocument) error

	// Delete methods
	DeleteOrganization(ctx context.Context, id string) error
	DeleteUser(ctx context.Context, id string) error
	DeleteProject(ctx context.Context, id string) error
	DeleteBoard(ctx context.Context, id string) error
	DeleteCard(ctx context.Context, id string) error

	// Initialize creates all collections if they don't exist
	InitializeCollections(ctx context.Context) error
}

type service struct {
	client     TypesenseClient
	memberRepo organization_member.Repository
}

// NewService creates a new search service using the TypesenseClient interface
func NewService(client TypesenseClient, memberRepo organization_member.Repository) Service {
	return &service{
		client:     client,
		memberRepo: memberRepo,
	}
}

// NewServiceFromRawClient creates a new search service from a raw Typesense client
// This is provided for backward compatibility
func NewServiceFromRawClient(client *typesense.Client, memberRepo organization_member.Repository) Service {
	return &service{
		client:     NewTypesenseClientFromRaw(client),
		memberRepo: memberRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "search.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "search"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

// InitializeCollections creates all search collections if they don't exist
func (s *service) InitializeCollections(ctx context.Context) error {
	ctx, span := s.startServiceSpan(ctx, "InitializeCollections")
	defer span.End()

	schemas := GetAllSchemas()
	for _, schema := range schemas {
		// Check if collection exists
		_, err := s.client.RetrieveCollection(ctx, schema.Name)
		if err == nil {
			// Collection exists, skip
			continue
		}

		// Create collection
		_, err = s.client.CreateCollection(ctx, schema)
		if err != nil {
			return fmt.Errorf("failed to create collection %s: %w", schema.Name, err)
		}
	}

	return nil
}

// getUserOrgIDs returns the organization IDs the user has access to
func (s *service) getUserOrgIDs(ctx context.Context, userID uuid.UUID) ([]string, error) {
	members, err := s.memberRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	orgIDs := make([]string, len(members))
	for i, m := range members {
		orgIDs[i] = m.OrganizationID.String()
	}
	return orgIDs, nil
}

// Search performs a multi-collection search with access control
func (s *service) Search(ctx context.Context, userID uuid.UUID, query string, scope *SearchScope, limit int) (*SearchResults, error) {
	ctx, span := s.startServiceSpan(ctx, "Search")
	span.SetAttributes(
		attribute.String("search.query", query),
		attribute.Int("search.limit", limit),
	)
	defer span.End()

	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	// Get user's accessible organization IDs for filtering
	orgIDs, err := s.getUserOrgIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	if len(orgIDs) == 0 {
		// User has no organizations, return empty results
		return &SearchResults{
			Results:    []*SearchResult{},
			TotalCount: 0,
			Query:      query,
		}, nil
	}

	// Build filter based on scope and access control
	orgFilter := fmt.Sprintf("organization_id:[%s]", strings.Join(orgIDs, ","))
	memberFilter := fmt.Sprintf("member_ids:[%s]", userID.String())
	userOrgFilter := fmt.Sprintf("organization_ids:[%s]", strings.Join(orgIDs, ","))

	// Apply scope filters if provided
	if scope != nil && scope.OrganizationID != "" {
		// Verify user has access to this org
		hasAccess := false
		for _, id := range orgIDs {
			if id == scope.OrganizationID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			return &SearchResults{
				Results:    []*SearchResult{},
				TotalCount: 0,
				Query:      query,
			}, nil
		}
		orgFilter = fmt.Sprintf("organization_id:=%s", scope.OrganizationID)
		memberFilter = fmt.Sprintf("member_ids:[%s] && id:=%s", userID.String(), scope.OrganizationID)
	}

	projectFilter := orgFilter
	if scope != nil && scope.ProjectID != "" {
		projectFilter = fmt.Sprintf("%s && project_id:=%s", orgFilter, scope.ProjectID)
	}

	// Build multi-search request
	searches := []api.MultiSearchCollectionParameters{
		{
			Collection: CollectionCards,
			Q:          pointer.String(query),
			QueryBy:    pointer.String("title,description"),
			FilterBy:   pointer.String(orgFilter),
			PerPage:    pointer.Int(limit),
		},
		{
			Collection: CollectionProjects,
			Q:          pointer.String(query),
			QueryBy:    pointer.String("name,key,description"),
			FilterBy:   pointer.String(orgFilter),
			PerPage:    pointer.Int(limit),
		},
		{
			Collection: CollectionBoards,
			Q:          pointer.String(query),
			QueryBy:    pointer.String("name,description"),
			FilterBy:   pointer.String(projectFilter),
			PerPage:    pointer.Int(limit),
		},
		{
			Collection: CollectionOrganizations,
			Q:          pointer.String(query),
			QueryBy:    pointer.String("name,slug,description"),
			FilterBy:   pointer.String(memberFilter),
			PerPage:    pointer.Int(limit),
		},
		{
			Collection: CollectionUsers,
			Q:          pointer.String(query),
			QueryBy:    pointer.String("username,email,display_name"),
			FilterBy:   pointer.String(userOrgFilter),
			PerPage:    pointer.Int(limit),
		},
	}

	// Execute multi-search
	params := &api.MultiSearchParams{}
	searchBody := api.MultiSearchSearchesParameter{
		Searches: searches,
	}
	resp, err := s.client.MultiSearch(ctx, params, searchBody)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Process results
	results := make([]*SearchResult, 0)
	totalCount := 0

	for i, searchResult := range resp.Results {
		if searchResult.Found == nil {
			continue
		}
		totalCount += *searchResult.Found

		if searchResult.Hits == nil {
			continue
		}

		for _, hit := range *searchResult.Hits {
			result := s.hitToSearchResult(hit, i)
			if result != nil {
				results = append(results, result)
			}
		}
	}

	return &SearchResults{
		Results:    results,
		TotalCount: totalCount,
		Query:      query,
	}, nil
}

func (s *service) hitToSearchResult(hit api.SearchResultHit, collectionIndex int) *SearchResult {
	if hit.Document == nil {
		return nil
	}

	doc := *hit.Document

	// Determine entity type based on collection index
	var entityType EntityType
	switch collectionIndex {
	case 0:
		entityType = EntityTypeCard
	case 1:
		entityType = EntityTypeProject
	case 2:
		entityType = EntityTypeBoard
	case 3:
		entityType = EntityTypeOrganization
	case 4:
		entityType = EntityTypeUser
	default:
		return nil
	}

	result := &SearchResult{
		Type:  entityType,
		ID:    getStringField(doc, "id"),
		Score: 0,
	}

	if hit.TextMatch != nil {
		result.Score = float64(*hit.TextMatch)
	}

	// Build highlight from matched fields
	if hit.Highlights != nil {
		var highlights []string
		for _, h := range *hit.Highlights {
			if h.Snippet != nil {
				highlights = append(highlights, *h.Snippet)
			}
		}
		result.Highlight = strings.Join(highlights, " ... ")
	}

	// Set fields based on entity type
	switch entityType {
	case EntityTypeCard:
		result.Title = getStringField(doc, "title")
		result.Description = getStringField(doc, "description")
		result.OrganizationID = getStringField(doc, "organization_id")
		result.OrganizationName = getStringField(doc, "organization_name")
		result.ProjectID = getStringField(doc, "project_id")
		result.ProjectName = getStringField(doc, "project_name")
		result.BoardID = getStringField(doc, "board_id")
		result.BoardName = getStringField(doc, "board_name")
		result.URL = fmt.Sprintf("/projects/%s/board/%s?card=%s", result.ProjectID, result.BoardID, result.ID)

	case EntityTypeProject:
		result.Title = getStringField(doc, "name")
		result.Description = getStringField(doc, "description")
		result.OrganizationID = getStringField(doc, "organization_id")
		result.OrganizationName = getStringField(doc, "organization_name")
		result.ProjectID = result.ID
		result.ProjectName = result.Title
		result.URL = fmt.Sprintf("/projects/%s", result.ID)

	case EntityTypeBoard:
		result.Title = getStringField(doc, "name")
		result.Description = getStringField(doc, "description")
		result.OrganizationID = getStringField(doc, "organization_id")
		result.OrganizationName = getStringField(doc, "organization_name")
		result.ProjectID = getStringField(doc, "project_id")
		result.ProjectName = getStringField(doc, "project_name")
		result.BoardID = result.ID
		result.BoardName = result.Title
		result.URL = fmt.Sprintf("/projects/%s/board/%s", result.ProjectID, result.BoardID)

	case EntityTypeOrganization:
		result.Title = getStringField(doc, "name")
		result.Description = getStringField(doc, "description")
		result.OrganizationID = result.ID
		result.OrganizationName = result.Title
		result.URL = fmt.Sprintf("/organizations/%s", result.ID)

	case EntityTypeUser:
		displayName := getStringField(doc, "display_name")
		if displayName == "" {
			displayName = getStringField(doc, "username")
		}
		result.Title = displayName
		result.Description = getStringField(doc, "email")
		result.URL = fmt.Sprintf("/users/%s", result.ID)
	}

	return result
}

func getStringField(doc map[string]interface{}, key string) string {
	if v, ok := doc[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// IndexOrganization indexes or updates an organization document
func (s *service) IndexOrganization(ctx context.Context, doc *OrganizationDocument) error {
	ctx, span := s.startServiceSpan(ctx, "IndexOrganization")
	span.SetAttributes(attribute.String("organization.id", doc.ID))
	defer span.End()

	_, err := s.client.UpsertDocument(ctx, CollectionOrganizations, doc)
	return err
}

// IndexUser indexes or updates a user document
func (s *service) IndexUser(ctx context.Context, doc *UserDocument) error {
	ctx, span := s.startServiceSpan(ctx, "IndexUser")
	span.SetAttributes(attribute.String("user.id", doc.ID))
	defer span.End()

	_, err := s.client.UpsertDocument(ctx, CollectionUsers, doc)
	return err
}

// IndexProject indexes or updates a project document
func (s *service) IndexProject(ctx context.Context, doc *ProjectDocument) error {
	ctx, span := s.startServiceSpan(ctx, "IndexProject")
	span.SetAttributes(attribute.String("project.id", doc.ID))
	defer span.End()

	_, err := s.client.UpsertDocument(ctx, CollectionProjects, doc)
	return err
}

// IndexBoard indexes or updates a board document
func (s *service) IndexBoard(ctx context.Context, doc *BoardDocument) error {
	ctx, span := s.startServiceSpan(ctx, "IndexBoard")
	span.SetAttributes(attribute.String("board.id", doc.ID))
	defer span.End()

	_, err := s.client.UpsertDocument(ctx, CollectionBoards, doc)
	return err
}

// IndexCard indexes or updates a card document
func (s *service) IndexCard(ctx context.Context, doc *CardDocument) error {
	ctx, span := s.startServiceSpan(ctx, "IndexCard")
	span.SetAttributes(attribute.String("card.id", doc.ID))
	defer span.End()

	_, err := s.client.UpsertDocument(ctx, CollectionCards, doc)
	return err
}

// DeleteOrganization removes an organization from the index
func (s *service) DeleteOrganization(ctx context.Context, id string) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteOrganization")
	span.SetAttributes(attribute.String("organization.id", id))
	defer span.End()

	_, err := s.client.DeleteDocument(ctx, CollectionOrganizations, id)
	return err
}

// DeleteUser removes a user from the index
func (s *service) DeleteUser(ctx context.Context, id string) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteUser")
	span.SetAttributes(attribute.String("user.id", id))
	defer span.End()

	_, err := s.client.DeleteDocument(ctx, CollectionUsers, id)
	return err
}

// DeleteProject removes a project from the index
func (s *service) DeleteProject(ctx context.Context, id string) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteProject")
	span.SetAttributes(attribute.String("project.id", id))
	defer span.End()

	_, err := s.client.DeleteDocument(ctx, CollectionProjects, id)
	return err
}

// DeleteBoard removes a board from the index
func (s *service) DeleteBoard(ctx context.Context, id string) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteBoard")
	span.SetAttributes(attribute.String("board.id", id))
	defer span.End()

	_, err := s.client.DeleteDocument(ctx, CollectionBoards, id)
	return err
}

// DeleteCard removes a card from the index
func (s *service) DeleteCard(ctx context.Context, id string) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteCard")
	span.SetAttributes(attribute.String("card.id", id))
	defer span.End()

	_, err := s.client.DeleteDocument(ctx, CollectionCards, id)
	return err
}
