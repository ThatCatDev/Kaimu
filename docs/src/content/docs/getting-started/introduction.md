---
title: Introduction
description: What is Pulse and why use it
---

Pulse is a modern, self-hosted project management tool designed for software teams. It provides a clean, fast interface for managing projects with Kanban boards, similar to tools like Jira or Linear.

## Why Pulse?

- **Self-Hosted**: Keep your data on your own infrastructure
- **Modern Stack**: Built with Go, GraphQL, Astro, and Svelte 5
- **Fast**: Optimized for speed and responsiveness
- **Flexible Authentication**: Works with any OIDC provider (Google, Okta, Azure AD, etc.)
- **Open Source**: Full access to the source code

## Tech Stack

### Backend
- **Go** - Fast, compiled backend
- **GraphQL** (gqlgen) - Type-safe API
- **PostgreSQL** - Reliable database
- **GORM** - Database ORM

### Frontend
- **Astro** - Static site generation with islands
- **Svelte 5** - Reactive UI components
- **TypeScript** - Type safety

## Core Features

### Organizations & Projects
Organize your work into organizations and projects. Each organization can have multiple projects, and users can be members of multiple organizations.

### Kanban Boards
Each project can have multiple Kanban boards with customizable columns. Drag and drop cards between columns to track progress.

### Cards & Tags
Create cards with descriptions, tags, and custom fields. Use tags to categorize and filter cards.

### Authentication
Secure authentication via OpenID Connect (OIDC). Works with:
- Google
- Okta
- Azure AD
- Auth0
- Keycloak
- Any OIDC-compliant provider

## Next Steps

- [Quick Start](/getting-started/quick-start/) - Get Pulse running in 5 minutes
- [Installation](/getting-started/installation/) - Detailed installation guide
- [Configuration](/configuration/environment-variables/) - Configure Pulse for your environment
