# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
- `go run cmd/main.go serve` - Start the GraphQL API server
- `make gql` - Regenerate GraphQL code from schema using gqlgen
- `make generate` - Generate all code (mocks + GraphQL)
- `make migrate-create name=<migration_name>` - Create new database migration

### Database
- `go run cmd/main.go migrate up` - Apply all pending database migrations
- `go run cmd/main.go migrate down` - Rollback one database migration
- `docker-compose up -d` - Start PostgreSQL locally

### Testing
- `go test ./...` - Run all tests

## Architecture

This is a GraphQL API built with Go, using:
- **gqlgen** for GraphQL server generation
- **GORM** with PostgreSQL for database operations
- **Cobra CLI** for command structure
- **OpenTelemetry** for distributed tracing
- **Prometheus** for metrics

### Key Structure
- `cmd/main.go` - Application entry point
- `internal/commands/` - CLI commands (serve, migrate up/down)
- `graph/` - GraphQL schema, resolvers, and generated code
- `internal/db/` - Database connection and tracing
- `internal/resolvers/` - GraphQL resolver implementations
- `db/migrations/` - Database migration files (PostgreSQL)

### GraphQL Schema
Schema files are in `graph/*.graphqls`. Currently has a single `helloWorld` query.

### Database
Uses PostgreSQL with GORM. Migrations use golang-migrate with SQL files.

### Configuration
Configuration loaded via `config/config.go` with development config in `config/config.dev.json`.

Environment variables:
- `DBHOST` - PostgreSQL host (default: localhost)
- `DBNAME` - Database name (default: pulse)
- `DBUSERNAME` - Database user (default: pulse)
- `DBPASSWORD` - Database password
- `DBPORT` - Database port (default: 5432)
- `DBSSL` - SSL mode (default: disable)

## Important Notes

### Code Generation
- GraphQL code is generated via gqlgen - modify `graph/*.graphqls` files and run `make gql`
- Always regenerate after schema changes

### Database Migrations
- Uses golang-migrate with SQL migration files in `db/migrations/`
- Sequential naming convention enforced by migrate tool
- Apply with `go run cmd/main.go migrate up`, rollback with `migrate down`

### Architecture Patterns

#### Dependency Injection in Resolvers
**IMPORTANT**: The `graph/resolver.go` should only contain **services**, not repositories.

```go
// CORRECT - only services in resolver
type Resolver struct {
    Config              config.Config
    AuthService         auth.Service
    OrganizationService organization.Service
    ProjectService      project.Service
}

// WRONG - do not add repositories directly
type Resolver struct {
    Config   config.Config
    UserRepo user.Repository  // DON'T DO THIS
}
```

If you need to access data in a resolver:
1. Add a method to an existing service
2. The service should encapsulate the repository access
3. Resolvers call service methods, never repositories directly

#### Service-Repository Pattern
- **Repositories**: Data access layer, CRUD operations only
- **Services**: Business logic, orchestrates repositories
- **Resolvers**: GraphQL layer, calls services only

Example flow:
```
Resolver -> Service -> Repository -> Database
```
