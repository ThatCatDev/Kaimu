# Kaimu

A modern project management tool for software teams, similar to Jira or Linear.

**kai** (change) + **mu** (nothing wasted)

## Overview

Kaimu helps software teams organize their work using Kanban boards, sprints, and agile workflows. It's built with a modern tech stack prioritizing developer experience and performance.

### Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Astro + Svelte 5, TypeScript |
| Backend | Go, GraphQL (gqlgen), GORM |
| Database | PostgreSQL |
| Auth | OIDC (OpenID Connect) |

## Core Concepts

### Hierarchy

```
Organization
    └── Project
            └── Board
                    ├── Columns (Todo, In Progress, Done, etc.)
                    │       └── Cards (tasks/issues)
                    └── Sprints
                            └── Cards (assigned to sprint)
```

### Organizations

Organizations are the top-level container. They represent a company, team, or group that shares projects and resources.

- **Members**: Users belong to organizations with assigned roles
- **Roles**: Owner, Admin, Member, Viewer (with granular permissions)
- **Invitations**: Invite users by email to join the organization

### Projects

Projects live within an organization and contain related work. Each project has:

- **Key**: A short identifier (e.g., "PROJ") used for card references
- **Boards**: One or more Kanban boards for organizing work
- **Tags**: Labels that can be applied to cards for categorization

### Boards

Boards visualize work using the Kanban methodology:

- **Columns**: Customizable workflow stages (Todo, In Progress, Review, Done)
- **WIP Limits**: Optional limits on cards per column
- **Done Column**: Mark a column as "done" for metrics tracking
- **Sprints**: Time-boxed iterations for agile workflows

### Cards

Cards represent units of work (tasks, bugs, stories):

- **Title & Description**: Rich text description with markdown support
- **Priority**: None, Low, Medium, High, Urgent
- **Assignee**: Team member responsible for the card
- **Due Date**: Optional deadline
- **Story Points**: Estimation for sprint planning
- **Tags**: Categorization labels
- **Sprints**: Can belong to multiple sprints (for carryover)

### Sprints

Sprints are time-boxed iterations for agile teams:

- **Status**: Future, Active, or Closed
- **Goal**: Sprint objective
- **Dates**: Start and end date
- **Cards**: Work items planned for the sprint
- **Metrics**: Burndown charts, velocity tracking

## Views

### Board View
Traditional Kanban board with drag-and-drop cards between columns.

### Planning View
Sprint planning interface showing:
- Active Sprint (expanded)
- Future Sprints (collapsed)
- Backlog (cards not in any sprint)
- Closed Sprints (for reference)

### Metrics View
Sprint analytics including:
- Burndown/Burnup charts
- Cumulative flow diagram
- Velocity metrics

## Database Schema

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              ORGANIZATION LAYER                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐         ┌─────────────────────┐         ┌──────────────┐  │
│  │    users     │────────<│organization_members │>────────│organizations │  │
│  ├──────────────┤         ├─────────────────────┤         ├──────────────┤  │
│  │ id           │         │ organization_id     │         │ id           │  │
│  │ username     │         │ user_id             │         │ name         │  │
│  │ email        │         │ role_id ────────────┼────┐    │ slug         │  │
│  │ display_name │         └─────────────────────┘    │    │ owner_id     │  │
│  │ avatar_url   │                                    │    └──────────────┘  │
│  │ password_hash│    ┌───────────┐    ┌─────────────┐│                      │
│  │ oidc_provider│    │permissions│───<│role_perms   │>───┌──────────┐       │
│  │ oidc_subject │    └───────────┘    └─────────────┘    │  roles   │       │
│  └──────────────┘                                        └──────────┘       │
│                                                                              │
│  ┌──────────────┐                                                           │
│  │ invitations  │  (pending org invites by email)                           │
│  └──────────────┘                                                           │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                               PROJECT LAYER                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐                              ┌──────────────┐              │
│  │organizations │─────────────────────────────<│   projects   │              │
│  └──────────────┘                              ├──────────────┤              │
│                                                │ id           │              │
│  ┌──────────────┐         ┌────────────────┐   │ org_id       │              │
│  │    users     │────────<│project_members │>──│ name         │              │
│  └──────────────┘         └────────────────┘   │ key          │              │
│                                                │ description  │              │
│                                                └──────────────┘              │
│                                                       │                      │
│                                                       │                      │
│                              ┌─────────────────┬──────┴──────┐               │
│                              ▼                 ▼             ▼               │
│                        ┌──────────┐      ┌──────────┐  ┌──────────┐          │
│                        │  boards  │      │   tags   │  │(future)  │          │
│                        └──────────┘      └──────────┘  │task_types│          │
│                                                        └──────────┘          │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                                BOARD LAYER                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│                        ┌──────────────────┐                                  │
│                        │      boards      │                                  │
│                        ├──────────────────┤                                  │
│                        │ id               │                                  │
│                        │ project_id       │                                  │
│                        │ name             │                                  │
│                        │ is_default       │                                  │
│                        └────────┬─────────┘                                  │
│                                 │                                            │
│              ┌──────────────────┼──────────────────┐                         │
│              ▼                  ▼                  ▼                         │
│       ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                  │
│       │board_columns│    │   sprints   │    │   cards     │                  │
│       ├─────────────┤    ├─────────────┤    ├─────────────┤                  │
│       │ id          │    │ id          │    │ id          │                  │
│       │ board_id    │    │ board_id    │    │ board_id    │                  │
│       │ name        │    │ name        │    │ column_id   │                  │
│       │ position    │    │ goal        │    │ title       │                  │
│       │ wip_limit   │    │ start_date  │    │ description │                  │
│       │ is_done     │    │ end_date    │    │ position    │                  │
│       │ is_backlog  │    │ status      │    │ priority    │                  │
│       │ color       │    │ position    │    │ assignee_id │                  │
│       └─────────────┘    └──────┬──────┘    │ due_date    │                  │
│              │                  │           │ story_points│                  │
│              │                  │           └──────┬──────┘                  │
│              │                  │                  │                         │
│              │                  │    ┌─────────────┼─────────────┐           │
│              │                  │    │             │             │           │
│              │                  ▼    ▼             ▼             ▼           │
│              │           ┌────────────────┐ ┌───────────┐ ┌───────────┐      │
│              │           │  card_sprints  │ │card_tags  │ │(assignee) │      │
│              │           │  (many-to-many)│ │(many-many)│ │  users    │      │
│              │           └────────────────┘ └───────────┘ └───────────┘      │
│              │                                                               │
│              └───────────────────────────────────────────────────────────────│
│                                      │                                       │
│                                      ▼                                       │
│                            ┌─────────────────┐                               │
│                            │ metrics_history │  (daily snapshots)            │
│                            ├─────────────────┤                               │
│                            │ sprint_id       │                               │
│                            │ recorded_date   │                               │
│                            │ total_cards     │                               │
│                            │ completed_cards │                               │
│                            │ story_points    │                               │
│                            │ column_snapshot │                               │
│                            └─────────────────┘                               │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                              AUDIT / TRACKING                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         audit_events                                 │    │
│  ├─────────────────────────────────────────────────────────────────────┤    │
│  │ id, occurred_at, actor_id, action, entity_type, entity_id           │    │
│  │ organization_id, project_id, board_id                               │    │
│  │ state_before (JSONB), state_after (JSONB), metadata (JSONB)         │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  Actions: created, updated, deleted, card_moved, card_assigned,             │
│           sprint_started, sprint_completed, member_invited, etc.            │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Entity Relationship Summary

| Relationship | Type | Description |
|--------------|------|-------------|
| Organization -> Projects | 1:N | An org has many projects |
| Organization -> Members | N:M | Users join orgs via organization_members |
| Project -> Boards | 1:N | A project has multiple boards |
| Project -> Tags | 1:N | Tags are scoped to a project |
| Board -> Columns | 1:N | A board has ordered columns |
| Board -> Cards | 1:N | Cards belong to a board |
| Board -> Sprints | 1:N | Sprints are board-level |
| Card -> Column | N:1 | A card is in one column |
| Card -> Sprints | N:M | Cards can be in multiple sprints |
| Card -> Tags | N:M | Cards can have multiple tags |
| Card -> Assignee | N:1 | A card has one assignee (optional) |
| Sprint -> Metrics | 1:N | Daily snapshots for charts |

## Role-Based Access Control (RBAC)

### System Roles

| Role | Description | Key Permissions |
|------|-------------|-----------------|
| **Owner** | Full access, cannot be removed | All permissions |
| **Admin** | Manage org and projects | All except delete org, manage roles |
| **Member** | Contribute to projects | View, create, edit cards |
| **Viewer** | Read-only access | View only |

### Permission Categories

- **Organization**: view, manage, delete, invite, remove members, manage roles
- **Project**: view, create, manage, delete, manage members
- **Board**: view, create, manage, delete
- **Card**: view, create, edit, move, delete, assign
- **Sprint**: view, manage

## Development

### Prerequisites

- Go 1.21+
- Bun (for frontend)
- PostgreSQL 15+
- Docker (optional, for local services)

### Quick Start

```bash
# Start PostgreSQL
docker-compose up -d

# Backend (from /backend)
go run cmd/main.go migrate up    # Run migrations
go run cmd/main.go serve         # Start server on :3000

# Frontend (from /frontend)
bun install
bun run dev                      # Start on :4321
```

### Project Structure

```
kaimu/
├── frontend/                # Astro + Svelte frontend
│   ├── src/
│   │   ├── components/      # Svelte components
│   │   │   ├── kanban/      # Board, columns, cards
│   │   │   ├── sprint/      # Sprint planning components
│   │   │   └── ui/          # Reusable UI components
│   │   ├── lib/
│   │   │   ├── api/         # API client functions
│   │   │   ├── graphql/     # GraphQL queries/mutations
│   │   │   └── stores/      # Svelte stores
│   │   ├── layouts/         # Page layouts
│   │   └── pages/           # Astro pages (routes)
│   └── e2e/                 # Playwright tests
│
├── backend/                 # Go GraphQL API
│   ├── cmd/                 # CLI entry point
│   ├── graph/               # GraphQL schema & generated code
│   ├── internal/
│   │   ├── commands/        # CLI commands (serve, migrate)
│   │   ├── db/              # Database connection
│   │   │   └── repositories/# Data access layer
│   │   ├── resolvers/       # GraphQL resolver implementations
│   │   └── services/        # Business logic
│   └── db/migrations/       # SQL migration files
│
└── docker-compose.yaml      # Local development services
```

### Common Commands

```bash
# Backend
make gql                          # Regenerate GraphQL code
make generate                     # Generate all code (mocks + GraphQL)
make migrate-create name=foo      # Create new migration
go test ./...                     # Run tests

# Frontend
bun run dev                       # Development server
bun run build                     # Production build
bun run codegen                   # Regenerate GraphQL types
bunx playwright test              # Run E2E tests
```

## License

MIT
