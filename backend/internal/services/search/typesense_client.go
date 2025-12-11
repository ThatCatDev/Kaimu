package search

//go:generate mockgen -source=typesense_client.go -destination=mocks/typesense_client_mock.go -package=mocks

import (
	"context"
	"fmt"

	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"
)

// TypesenseClient defines the interface for Typesense operations used by the search service.
// This interface allows for easier testing by enabling mock implementations.
type TypesenseClient interface {
	// Collection operations
	RetrieveCollection(ctx context.Context, name string) (*api.CollectionResponse, error)
	CreateCollection(ctx context.Context, schema *api.CollectionSchema) (*api.CollectionResponse, error)

	// Document operations
	UpsertDocument(ctx context.Context, collection string, document interface{}) (map[string]interface{}, error)
	DeleteDocument(ctx context.Context, collection string, id string) (map[string]interface{}, error)

	// Search operations
	MultiSearch(ctx context.Context, params *api.MultiSearchParams, searches api.MultiSearchSearchesParameter) (*api.MultiSearchResult, error)
}

// typesenseClientImpl wraps the actual Typesense client to implement TypesenseClient interface
type typesenseClientImpl struct {
	client *typesense.Client
}

// NewTypesenseClient creates a new TypesenseClient from config
func NewTypesenseClient(cfg config.TypesenseConfig) (TypesenseClient, error) {
	client := typesense.NewClient(
		typesense.WithServer(fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)),
		typesense.WithAPIKey(cfg.APIKey),
	)
	return &typesenseClientImpl{client: client}, nil
}

// NewTypesenseClientFromRaw creates a TypesenseClient from an existing raw client
func NewTypesenseClientFromRaw(client *typesense.Client) TypesenseClient {
	return &typesenseClientImpl{client: client}
}

func (c *typesenseClientImpl) RetrieveCollection(ctx context.Context, name string) (*api.CollectionResponse, error) {
	return c.client.Collection(name).Retrieve(ctx)
}

func (c *typesenseClientImpl) CreateCollection(ctx context.Context, schema *api.CollectionSchema) (*api.CollectionResponse, error) {
	return c.client.Collections().Create(ctx, schema)
}

func (c *typesenseClientImpl) UpsertDocument(ctx context.Context, collection string, document interface{}) (map[string]interface{}, error) {
	return c.client.Collection(collection).Documents().Upsert(ctx, document)
}

func (c *typesenseClientImpl) DeleteDocument(ctx context.Context, collection string, id string) (map[string]interface{}, error) {
	return c.client.Collection(collection).Document(id).Delete(ctx)
}

func (c *typesenseClientImpl) MultiSearch(ctx context.Context, params *api.MultiSearchParams, searches api.MultiSearchSearchesParameter) (*api.MultiSearchResult, error) {
	return c.client.MultiSearch.Perform(ctx, params, searches)
}
