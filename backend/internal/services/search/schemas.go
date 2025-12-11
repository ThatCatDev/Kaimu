package search

import (
	"github.com/typesense/typesense-go/v2/typesense/api"
)

// Collection names
const (
	CollectionOrganizations = "organizations"
	CollectionUsers         = "users"
	CollectionProjects      = "projects"
	CollectionBoards        = "boards"
	CollectionCards         = "cards"
)

// Ptr returns a pointer to the value
func Ptr[T any](v T) *T {
	return &v
}

// GetOrganizationSchema returns the Typesense schema for organizations
func GetOrganizationSchema() *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: CollectionOrganizations,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "slug", Type: "string"},
			{Name: "description", Type: "string", Optional: Ptr(true)},
			{Name: "owner_id", Type: "string"},
			{Name: "member_ids", Type: "string[]"}, // For access control
			{Name: "created_at", Type: "int64"},
			{Name: "updated_at", Type: "int64"},
		},
		DefaultSortingField: Ptr("updated_at"),
	}
}

// GetUserSchema returns the Typesense schema for users
func GetUserSchema() *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: CollectionUsers,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "username", Type: "string"},
			{Name: "email", Type: "string", Optional: Ptr(true)},
			{Name: "display_name", Type: "string", Optional: Ptr(true)},
			{Name: "organization_ids", Type: "string[]"}, // For access control
			{Name: "created_at", Type: "int64"},
		},
		DefaultSortingField: Ptr("created_at"),
	}
}

// GetProjectSchema returns the Typesense schema for projects
func GetProjectSchema() *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: CollectionProjects,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "key", Type: "string"},
			{Name: "description", Type: "string", Optional: Ptr(true)},
			{Name: "organization_id", Type: "string"},
			{Name: "organization_name", Type: "string"},
			{Name: "organization_slug", Type: "string"},
			{Name: "created_at", Type: "int64"},
			{Name: "updated_at", Type: "int64"},
		},
		DefaultSortingField: Ptr("updated_at"),
	}
}

// GetBoardSchema returns the Typesense schema for boards
func GetBoardSchema() *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: CollectionBoards,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "description", Type: "string", Optional: Ptr(true)},
			{Name: "is_default", Type: "bool"},
			{Name: "project_id", Type: "string"},
			{Name: "project_name", Type: "string"},
			{Name: "project_key", Type: "string"},
			{Name: "organization_id", Type: "string"},
			{Name: "organization_name", Type: "string"},
			{Name: "organization_slug", Type: "string"},
			{Name: "created_at", Type: "int64"},
			{Name: "updated_at", Type: "int64"},
		},
		DefaultSortingField: Ptr("updated_at"),
	}
}

// GetCardSchema returns the Typesense schema for cards
func GetCardSchema() *api.CollectionSchema {
	return &api.CollectionSchema{
		Name: CollectionCards,
		Fields: []api.Field{
			{Name: "id", Type: "string"},
			{Name: "title", Type: "string"},
			{Name: "description", Type: "string", Optional: Ptr(true)},
			{Name: "priority", Type: "string"},
			{Name: "board_id", Type: "string"},
			{Name: "board_name", Type: "string"},
			{Name: "project_id", Type: "string"},
			{Name: "project_name", Type: "string"},
			{Name: "project_key", Type: "string"},
			{Name: "organization_id", Type: "string"},
			{Name: "organization_name", Type: "string"},
			{Name: "organization_slug", Type: "string"},
			{Name: "assignee_id", Type: "string", Optional: Ptr(true)},
			{Name: "assignee_name", Type: "string", Optional: Ptr(true)},
			{Name: "tags", Type: "string[]", Optional: Ptr(true)},
			{Name: "due_date", Type: "int64", Optional: Ptr(true)},
			{Name: "created_at", Type: "int64"},
			{Name: "updated_at", Type: "int64"},
		},
		DefaultSortingField: Ptr("updated_at"),
	}
}

// GetAllSchemas returns all collection schemas
func GetAllSchemas() []*api.CollectionSchema {
	return []*api.CollectionSchema{
		GetOrganizationSchema(),
		GetUserSchema(),
		GetProjectSchema(),
		GetBoardSchema(),
		GetCardSchema(),
	}
}
