---
title: Introduction
description: What is Kaimu and why use it
---

Kaimu is a modern, self-hosted project management tool designed for software teams. It provides a clean, fast interface for managing projects with Kanban boards, sprints, and agile workflows - similar to tools like Jira or Linear.

**kai** (change) + **mu** (nothing wasted)

## Why Kaimu?

- **Self-Hosted**: Keep your data on your own infrastructure
- **Modern Stack**: Built with Go, GraphQL, Astro, and Svelte 5
- **Agile Ready**: Kanban boards, sprints, planning views, and burndown charts
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
Organize your work into organizations and projects. Each organization can have multiple projects, and users can be members of multiple organizations with role-based access control.

### Kanban Boards
Each project can have multiple Kanban boards with customizable columns, WIP limits, and drag-and-drop cards.

### Sprints & Planning
Use sprints for time-boxed iterations. The Planning view helps you manage your backlog, assign cards to sprints, and track progress.

### Cards & Tags
Create cards with rich text descriptions, priorities, assignees, story points, due dates, and tags. Track everything your team needs.

### Metrics & Charts
Visualize progress with burndown charts, cumulative flow diagrams, and velocity tracking.

### Authentication
Secure authentication via OpenID Connect (OIDC). Works with:
- Google
- Okta
- Azure AD
- Auth0
- Keycloak
- Any OIDC-compliant provider

## Next Steps

- [Quick Start](/getting-started/quick-start/) - Get Kaimu running in 5 minutes
- [Core Concepts](/usage/concepts/) - Learn about organizations, projects, boards, and sprints
- [Installation](/getting-started/installation/) - Detailed installation guide
- [Configuration](/configuration/environment-variables/) - Configure Kaimu for your environment
