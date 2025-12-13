---
title: Contributing
description: How to contribute to Kaimu
---

We welcome contributions to Kaimu! This guide will help you get started.

## Getting Started

### Prerequisites

- Go 1.21+
- Bun (Node.js runtime)
- Docker and Docker Compose
- Git

### Development Setup

1. Fork and clone the repository:

```bash
git clone https://github.com/YOUR_USERNAME/kaimu.git
cd kaimu
```

2. Start the development environment:

```bash
docker compose up -d
```

3. Set up the backend:

```bash
cd backend
go mod download
go run cmd/main.go migrate up
```

4. Set up the frontend:

```bash
cd frontend
bun install
```

5. Start development servers:

```bash
# Terminal 1: Backend
cd backend
air  # or: go run cmd/main.go serve

# Terminal 2: Frontend
cd frontend
bun run dev
```

## Project Structure

```
kaimu/
├── backend/          # Go GraphQL API
├── frontend/         # Astro + Svelte frontend
├── docs/             # Documentation (Starlight)
├── dex/              # Dex configuration for local OIDC
└── docker-compose.yaml
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Changes

- Follow the existing code style
- Add tests for new functionality
- Update documentation if needed

### 3. Test Your Changes

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
bun run test

# E2E tests
bun run test:e2e
```

### 4. Commit Your Changes

Use conventional commit messages:

```bash
git commit -m "feat: add new feature"
git commit -m "fix: resolve bug in component"
git commit -m "docs: update installation guide"
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

### 5. Push and Create a Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Code Guidelines

### Go (Backend)

- Follow standard Go conventions
- Use `gofmt` for formatting
- Write tests for new functionality
- Use meaningful variable and function names

```go
// Good
func (s *service) CreateOrganization(ctx context.Context, name string) (*Organization, error) {
    // ...
}

// Avoid
func (s *service) Create(ctx context.Context, n string) (*Organization, error) {
    // ...
}
```

### TypeScript/Svelte (Frontend)

- Use TypeScript for type safety
- Follow Svelte 5 patterns (runes)
- Keep components focused and small
- Use meaningful component names

```svelte
<!-- Good: Clear, typed props -->
<script lang="ts">
  interface Props {
    title: string;
    onClose: () => void;
  }
  let { title, onClose }: Props = $props();
</script>
```

### GraphQL

- Use meaningful type and field names
- Add descriptions to types and fields
- Follow the existing schema patterns

```graphql
"""
Organization represents a team or company
"""
type Organization {
  """
  Unique identifier
  """
  id: ID!

  """
  Display name of the organization
  """
  name: String!
}
```

## Testing

### Backend Tests

```bash
cd backend
go test ./...                    # Run all tests
go test ./internal/services/...  # Run specific package
go test -v ./...                 # Verbose output
go test -cover ./...             # With coverage
```

### Frontend Tests

```bash
cd frontend
bun run test           # Unit tests
bun run test:e2e       # E2E tests (Playwright)
bun run test:e2e:ui    # E2E with UI
```

## Making Database Changes

### 1. Create a Migration

```bash
cd backend
make migrate-create name=add_new_table
```

### 2. Write the Migration

Edit the generated files in `backend/db/migrations/`:

```sql
-- XXXXXX_add_new_table.up.sql
CREATE TABLE new_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- XXXXXX_add_new_table.down.sql
DROP TABLE IF EXISTS new_table;
```

### 3. Run the Migration

```bash
go run cmd/main.go migrate up
```

## Making GraphQL Changes

### Backend

1. Edit schema files in `backend/graph/`:
   - `schema.graphqls` - Queries and mutations
   - `types.graphqls` - Type definitions

2. Regenerate code:
```bash
make gql
```

3. Implement resolvers in `backend/internal/resolvers/`

### Frontend

1. Edit queries in `frontend/src/lib/graphql/*.graphql`

2. Regenerate types:
```bash
bun run codegen
```

## Documentation

### Running Docs Locally

```bash
cd docs
bun install
bun run dev
```

### Adding Documentation

1. Create a new `.md` file in `docs/src/content/docs/`
2. Add frontmatter:

```markdown
---
title: Page Title
description: Page description
---

Content here...
```

3. Update `astro.config.mjs` if needed for navigation

## Getting Help

- Open an issue for bugs or feature requests
- Join discussions in GitHub Discussions
- Ask questions in pull requests

## Code of Conduct

Be respectful and inclusive. We're all here to build something great together.

## License

Kaimu is open source software. By contributing, you agree that your contributions will be licensed under the same license.
