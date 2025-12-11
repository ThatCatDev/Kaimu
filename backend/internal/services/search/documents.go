package search

import "time"

// EntityType represents the type of searchable entity
type EntityType string

const (
	EntityTypeOrganization EntityType = "organization"
	EntityTypeUser         EntityType = "user"
	EntityTypeProject      EntityType = "project"
	EntityTypeBoard        EntityType = "board"
	EntityTypeCard         EntityType = "card"
)

// OrganizationDocument represents an organization in the search index
type OrganizationDocument struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	OwnerID     string   `json:"owner_id"`
	MemberIDs   []string `json:"member_ids"` // For access control filtering
	CreatedAt   int64    `json:"created_at"` // Unix timestamp
	UpdatedAt   int64    `json:"updated_at"`
}

// UserDocument represents a user in the search index
type UserDocument struct {
	ID              string   `json:"id"`
	Username        string   `json:"username"`
	Email           string   `json:"email"`
	DisplayName     string   `json:"display_name"`
	OrganizationIDs []string `json:"organization_ids"` // For access control filtering
	CreatedAt       int64    `json:"created_at"`
}

// ProjectDocument represents a project in the search index
type ProjectDocument struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Key              string `json:"key"`
	Description      string `json:"description"`
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	OrganizationSlug string `json:"organization_slug"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
}

// BoardDocument represents a board in the search index
type BoardDocument struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	IsDefault        bool   `json:"is_default"`
	ProjectID        string `json:"project_id"`
	ProjectName      string `json:"project_name"`
	ProjectKey       string `json:"project_key"`
	OrganizationID   string `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	OrganizationSlug string `json:"organization_slug"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
}

// CardDocument represents a card in the search index
type CardDocument struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Priority         string   `json:"priority"`
	BoardID          string   `json:"board_id"`
	BoardName        string   `json:"board_name"`
	ProjectID        string   `json:"project_id"`
	ProjectName      string   `json:"project_name"`
	ProjectKey       string   `json:"project_key"`
	OrganizationID   string   `json:"organization_id"`
	OrganizationName string   `json:"organization_name"`
	OrganizationSlug string   `json:"organization_slug"`
	AssigneeID       string   `json:"assignee_id"`
	AssigneeName     string   `json:"assignee_name"`
	Tags             []string `json:"tags"`
	DueDate          int64    `json:"due_date"` // Unix timestamp, 0 if not set
	CreatedAt        int64    `json:"created_at"`
	UpdatedAt        int64    `json:"updated_at"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Type             EntityType `json:"type"`
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description,omitempty"`
	Highlight        string     `json:"highlight"`
	OrganizationID   string     `json:"organization_id"`
	OrganizationName string     `json:"organization_name"`
	ProjectID        string     `json:"project_id,omitempty"`
	ProjectName      string     `json:"project_name,omitempty"`
	BoardID          string     `json:"board_id,omitempty"`
	BoardName        string     `json:"board_name,omitempty"`
	URL              string     `json:"url"`
	Score            float64    `json:"score"`
}

// SearchResults represents the search response
type SearchResults struct {
	Results    []*SearchResult `json:"results"`
	TotalCount int             `json:"total_count"`
	Query      string          `json:"query"`
}

// SearchScope defines the context for filtering search results
type SearchScope struct {
	OrganizationID string
	ProjectID      string
}

// Helper function to convert time to unix timestamp
func ToUnixTimestamp(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

// Helper function to convert *time.Time to unix timestamp
func ToUnixTimestampPtr(t *time.Time) int64 {
	if t == nil || t.IsZero() {
		return 0
	}
	return t.Unix()
}
