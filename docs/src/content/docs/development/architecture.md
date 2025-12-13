---
title: Architecture
description: Technical architecture of Kaimu
---

This document describes the technical architecture of Kaimu.

## Overview

Kaimu is built as a modern web application with a clear separation between frontend and backend:

```mermaid
graph TB
    subgraph Frontend
        A1[Astro SSG]
        A2[Svelte 5 Components]
        A3[GraphQL Client]
    end

    subgraph Backend
        B1[gqlgen GraphQL]
        B2[Services]
        B3[Repositories]
    end

    subgraph Database
        C1[(PostgreSQL)]
    end

    A1 --> A2
    A2 --> A3
    A3 -->|GraphQL| B1
    B1 --> B2
    B2 --> B3
    B3 -->|SQL| C1

    style A1 fill:#7C3AED,color:#fff
    style A2 fill:#F97316,color:#fff
    style A3 fill:#0EA5E9,color:#fff
    style B1 fill:#10B981,color:#fff
    style B2 fill:#10B981,color:#fff
    style B3 fill:#10B981,color:#fff
    style C1 fill:#3B82F6,color:#fff
```

## Frontend Architecture

### Technology Stack

- **Astro** - Static site generation with islands architecture
- **Svelte 5** - Reactive components with runes (`$state`, `$derived`, `$effect`)
- **TypeScript** - Type safety
- **Bits UI** - Accessible UI components

### Directory Structure

```
frontend/
├── src/
│   ├── components/     # Svelte components
│   ├── layouts/        # Astro layouts
│   ├── pages/          # Astro pages (file-based routing)
│   ├── lib/
│   │   ├── api/        # API client functions
│   │   └── graphql/    # GraphQL queries and generated types
│   └── styles/         # Global styles
├── astro.config.mjs
└── package.json
```

### Key Patterns

#### Islands Architecture

Astro renders pages as static HTML, with Svelte components hydrated as interactive "islands":

```astro
---
// src/pages/dashboard.astro
import Layout from '../layouts/Layout.astro';
import Dashboard from '../components/Dashboard.svelte';
---

<Layout>
  <Dashboard client:load />
</Layout>
```

#### Svelte 5 Runes

Components use Svelte 5's runes for state management:

```svelte
<script lang="ts">
  let count = $state(0);
  let doubled = $derived(count * 2);

  $effect(() => {
    console.log('Count changed:', count);
  });
</script>
```

## Backend Architecture

### Technology Stack

- **Go** - Compiled, performant backend
- **gqlgen** - Code-first GraphQL server
- **GORM** - Database ORM
- **PostgreSQL** - Relational database

### Directory Structure

```
backend/
├── cmd/
│   └── main.go           # Entry point
├── config/               # Configuration
├── graph/
│   ├── schema.graphqls   # GraphQL schema
│   ├── types.graphqls    # GraphQL types
│   ├── generated/        # Generated code
│   └── resolver.go       # Resolver dependency injection
├── internal/
│   ├── commands/         # CLI commands
│   ├── db/
│   │   └── repositories/ # Data access layer
│   ├── resolvers/        # GraphQL resolver implementations
│   └── services/         # Business logic
├── http/
│   ├── handlers/         # HTTP handlers
│   └── middleware/       # HTTP middleware
└── db/
    └── migrations/       # SQL migrations
```

### Layer Architecture

```mermaid
graph TB
    A[HTTP Layer<br/>handlers, middleware, routing] --> B
    B[GraphQL Layer<br/>resolvers, schema, types] --> C
    C[Service Layer<br/>business logic] --> D
    D[Repository Layer<br/>data access] --> E
    E[(Database<br/>PostgreSQL)]

    style A fill:#EC4899,color:#fff
    style B fill:#8B5CF6,color:#fff
    style C fill:#3B82F6,color:#fff
    style D fill:#10B981,color:#fff
    style E fill:#F59E0B,color:#fff
```

### Key Patterns

#### Dependency Injection

Services are injected into the GraphQL resolver:

```go
// graph/resolver.go
type Resolver struct {
    Config              config.Config
    AuthService         auth.Service
    OrganizationService organization.Service
    ProjectService      project.Service
    // ...
}
```

#### Repository Pattern

Data access is abstracted behind repository interfaces:

```go
// internal/db/repositories/user/user_repository.go
type Repository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
}
```

#### Service Pattern

Business logic lives in services:

```go
// internal/services/organization/organization_service.go
type Service interface {
    Create(ctx context.Context, name, description string, ownerID uuid.UUID) (*Organization, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
    GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*Organization, error)
}
```

## Authentication Flow

### OIDC Authentication

```mermaid
sequenceDiagram
    participant B as Browser
    participant F as Frontend
    participant BE as Backend
    participant O as OIDC Provider

    B->>F: Click Login
    F->>B: Redirect to /auth/oidc/...
    B->>BE: GET /auth/oidc/{provider}/authorize
    BE->>B: Redirect to OIDC Provider
    B->>O: Authenticate with Provider
    O->>B: Redirect with auth code
    B->>BE: GET /auth/oidc/{provider}/callback?code=...
    BE->>O: Exchange code for tokens
    O->>BE: ID Token + Access Token
    BE->>BE: Create/link user, generate token pair
    BE->>B: Set access + refresh cookies, redirect
    B->>F: Load dashboard page
```

### Token Refresh Flow

```mermaid
sequenceDiagram
    participant F as Frontend
    participant BE as Backend
    participant DB as Database

    Note over F: Access token expires in 30s
    F->>BE: POST /graphql (refreshToken mutation)
    BE->>DB: Validate refresh token
    alt Token valid and not revoked
        BE->>DB: Revoke old refresh token
        BE->>DB: Create new refresh token
        BE->>F: Set new access + refresh cookies
        F->>BE: Retry original request
    else Token invalid or revoked
        alt Reuse detected
            BE->>DB: Revoke ALL user tokens
        end
        BE->>F: 401 Unauthorized
        F->>F: Redirect to login
    end
```

## Database Schema

### Entity Relationships

```mermaid
erDiagram
    users ||--o{ organization_members : "belongs to"
    organizations ||--o{ organization_members : "has"
    organizations ||--o{ projects : "contains"
    organizations ||--o{ invitations : "has"

    projects ||--o{ boards : "has"
    projects ||--o{ tags : "has"
    projects ||--o{ project_members : "has"
    users ||--o{ project_members : "belongs to"

    boards ||--o{ board_columns : "has"
    boards ||--o{ cards : "contains"
    boards ||--o{ sprints : "has"

    board_columns ||--o{ cards : "contains"

    cards ||--o{ card_tags : "has"
    tags ||--o{ card_tags : "tagged with"
    cards ||--o{ card_sprints : "assigned to"
    sprints ||--o{ card_sprints : "contains"
    sprints ||--o{ metrics_history : "tracked by"

    users ||--o{ cards : "assigned to"
    users ||--o{ oidc_identities : "has"
    users ||--o{ email_verification_tokens : "has"
    users ||--o{ refresh_tokens : "has"

    roles ||--o{ role_permissions : "has"
    permissions ||--o{ role_permissions : "granted by"
    roles ||--o{ organization_members : "assigned to"
    roles ||--o{ project_members : "assigned to"

    users {
        uuid id PK
        string username
        string email
        string display_name
        string avatar_url
        string password_hash
        string oidc_provider
        string oidc_subject
        boolean email_verified
        timestamp created_at
    }

    organizations {
        uuid id PK
        string name
        string slug UK
        string description
        uuid owner_id FK
        timestamp created_at
    }

    organization_members {
        uuid id PK
        uuid organization_id FK
        uuid user_id FK
        uuid role_id FK
        timestamp created_at
    }

    projects {
        uuid id PK
        uuid organization_id FK
        string name
        string key
        string description
        timestamp created_at
    }

    boards {
        uuid id PK
        uuid project_id FK
        string name
        string description
        boolean is_default
        timestamp created_at
    }

    board_columns {
        uuid id PK
        uuid board_id FK
        string name
        int position
        string color
        int wip_limit
        boolean is_done
        boolean is_backlog
    }

    cards {
        uuid id PK
        uuid board_id FK
        uuid column_id FK
        string title
        text description
        float position
        enum priority
        uuid assignee_id FK
        timestamp due_date
        int story_points
        timestamp created_at
    }

    tags {
        uuid id PK
        uuid project_id FK
        string name
        string color
    }

    card_tags {
        uuid id PK
        uuid card_id FK
        uuid tag_id FK
    }

    sprints {
        uuid id PK
        uuid board_id FK
        string name
        text goal
        timestamp start_date
        timestamp end_date
        enum status
        int position
    }

    card_sprints {
        uuid id PK
        uuid card_id FK
        uuid sprint_id FK
        timestamp added_at
    }

    metrics_history {
        uuid id PK
        uuid sprint_id FK
        date recorded_date
        int total_cards
        int completed_cards
        int total_story_points
        int completed_story_points
        jsonb column_snapshot
    }

    roles {
        uuid id PK
        uuid organization_id FK
        string name
        string description
        boolean is_system
        string scope
    }

    permissions {
        uuid id PK
        string code UK
        string name
        string description
        string resource_type
    }

    role_permissions {
        uuid id PK
        uuid role_id FK
        uuid permission_id FK
    }

    invitations {
        uuid id PK
        uuid organization_id FK
        string email
        uuid role_id FK
        uuid invited_by FK
        string token UK
        timestamp expires_at
        timestamp accepted_at
    }

    audit_events {
        uuid id PK
        timestamp occurred_at
        uuid actor_id FK
        enum action
        enum entity_type
        uuid entity_id
        uuid organization_id FK
        uuid project_id FK
        uuid board_id FK
        jsonb state_before
        jsonb state_after
        jsonb metadata
    }

    project_members {
        uuid id PK
        uuid project_id FK
        uuid user_id FK
        uuid role_id FK
        timestamp created_at
    }

    oidc_identities {
        uuid id PK
        uuid user_id FK
        string issuer
        string subject
        string email
        boolean email_verified
        timestamp created_at
        timestamp updated_at
    }

    email_verification_tokens {
        uuid id PK
        uuid user_id FK
        string token UK
        string email
        timestamp expires_at
        timestamp used_at
        timestamp created_at
    }

    refresh_tokens {
        uuid id PK
        uuid user_id FK
        string token_hash
        timestamp expires_at
        timestamp created_at
        timestamp revoked_at
        uuid replaced_by FK
        text user_agent
        string ip_address
    }
```

## Code Generation

### GraphQL (Backend)

After modifying `graph/*.graphqls`:

```bash
cd backend
make gql
```

This generates:
- `graph/generated/generated.go` - Server code
- `graph/model/models_gen.go` - Go types

### GraphQL (Frontend)

After modifying `src/lib/graphql/*.graphql`:

```bash
cd frontend
bun run codegen
```

This generates:
- `src/lib/graphql/generated.ts` - TypeScript types and functions
