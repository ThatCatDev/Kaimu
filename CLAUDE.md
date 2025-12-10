# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Kaimu** is a project management tool for software teams, similar to Jira or Linear.

_kai (change) + mu (nothing wasted)_

### Tech Stack
- **Frontend**: Astro + Svelte 5, TypeScript, Vite (via Bun)
- **Backend**: Go with gqlgen (GraphQL), GORM, PostgreSQL
- **Authentication**: OIDC (OpenID Connect) for any identity provider
- **API**: GraphQL

## Project Structure

```
pulse/
├── frontend/          # Astro + Svelte frontend
│   ├── src/
│   │   ├── components/   # Svelte components
│   │   ├── layouts/      # Astro layouts
│   │   └── pages/        # Astro pages
│   └── astro.config.mjs
├── backend/           # Go GraphQL API
│   ├── cmd/              # CLI entry point
│   ├── graph/            # GraphQL schema and resolvers
│   ├── internal/         # Application code
│   │   ├── db/           # Database connection
│   │   ├── resolvers/    # Resolver implementations
│   │   └── commands/     # CLI commands
│   └── db/migrations/    # PostgreSQL migrations
└── docker-compose.yaml   # Local development services
```

## Core Features to Implement

### Authentication & Authorization
- **OIDC Support**: Integrate with any OpenID Connect identity provider (Okta, Auth0, Keycloak, Google, etc.)
- **User Management**: User profiles linked to OIDC identities
- **Roles & Permissions**: Admin, Project Manager, Developer, Viewer roles with granular permissions

### Organizational Structure
- **Organizations**: Top-level container for teams and projects
- **Teams**: Groups of users that can be assigned to projects
- **Projects**: Containers for work items with team assignments

### Work Items
- **Task Types**: Epic, Story, Task, Bug, Subtask (hierarchical)
- **Subtask Support**: Tasks can have unlimited nested subtasks
- **Custom Fields**: Extensible fields per task type
- **Relationships**: Blocks, blocked by, relates to, duplicates

### Agile Features
- **Kanban Boards**: Customizable columns, WIP limits, swimlanes
- **Backlogs**: Prioritized list of work items
- **Sprints**: Time-boxed iterations with capacity planning
- **Story Points**: Estimation support

### Time & Progress
- **Time Tracking**: Log time against tasks
- **Burndown Charts**: Sprint progress visualization
- **Burnup Charts**: Scope and completion tracking
- **Velocity Metrics**: Team performance over time

### Reporting & Analytics
- **Dashboards**: Customizable project dashboards
- **Reports**: Sprint reports, velocity, cycle time, lead time
- **Export**: CSV/PDF export capabilities

## Development Commands

### Backend (from `/backend`)
```bash
docker-compose up -d              # Start PostgreSQL
go run cmd/main.go serve          # Start GraphQL server (port 3000)
go run cmd/main.go migrate up     # Run migrations
make gql                          # Regenerate GraphQL code
```

### Frontend (from `/frontend`)
```bash
bun run dev      # Start dev server
bun run build    # Production build
bun run preview  # Preview build
```

## Database Schema Guidelines

### Core Entities
- `users` - User profiles (linked to OIDC subject)
- `organizations` - Top-level org
- `teams` - Team within org
- `team_members` - User-team membership with role
- `projects` - Project within org
- `project_teams` - Team-project assignment
- `task_types` - Configurable task types per org
- `tasks` - Work items (self-referential for subtasks)
- `task_comments` - Comments on tasks
- `task_attachments` - File attachments
- `sprints` - Sprint definitions
- `sprint_tasks` - Task-sprint assignments
- `time_entries` - Time logged against tasks
- `boards` - Kanban board definitions
- `board_columns` - Board column configuration

### Key Relationships
- Tasks have a `parent_id` for subtask hierarchy
- Tasks belong to a project and optionally a sprint
- Users belong to organizations through team membership
- Projects are accessible to assigned teams

## GraphQL Schema Guidelines

When adding new types/queries:
1. Define types in `graph/types.graphqls`
2. Define queries/mutations in `graph/schema.graphqls`
3. Run `make gql` to regenerate
4. Implement resolvers in `graph/schema.resolvers.go`
5. Add business logic in `internal/resolvers/`

## Authentication Flow

1. Frontend redirects to OIDC provider
2. Provider authenticates and redirects back with code
3. Backend exchanges code for tokens
4. Backend creates/updates user from OIDC claims
5. Backend issues session token (JWT or session cookie)
6. Frontend includes token in GraphQL requests

## Current Status

- [x] Project structure setup
- [x] Backend GraphQL skeleton with PostgreSQL
- [x] Frontend Astro + Svelte setup
- [ ] Database schema and migrations
- [ ] OIDC authentication
- [ ] User and organization management
- [ ] Task CRUD operations
- [ ] Kanban board UI
- [ ] Sprint management
- [ ] Time tracking
- [ ] Reporting and charts
