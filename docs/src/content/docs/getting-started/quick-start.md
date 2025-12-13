---
title: Quick Start
description: Get Kaimu running in 5 minutes with Docker Compose
---

Get Kaimu running locally in just a few minutes using Docker Compose.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [Git](https://git-scm.com/)

## Steps

### 1. Clone the Repository

```bash
git clone https://github.com/ThatCatDev/Kaimu.git
cd Kaimu
```

### 2. Start the Services

```bash
docker compose up -d
```

This starts:
- **PostgreSQL** - Database (port 5432)
- **Dex** - OIDC provider for local development (port 5556)
- **Backend** - Go GraphQL API (port 3000)
- **Frontend** - Astro/Svelte app (port 4321)

### 3. Run Database Migrations

```bash
docker compose exec backend go run cmd/main.go migrate up
```

### 4. Access Kaimu

Open [http://localhost:4321](http://localhost:4321) in your browser.

### 5. Login

Click "Continue with Dex" and use these test credentials:

| Email | Password |
|-------|----------|
| `admin@kaimu.local` | `password` |
| `user@kaimu.local` | `password` |

## What's Next?

- Create your first organization
- Create a project within the organization
- Set up a Kanban board
- Add cards and start tracking work

## Stopping Kaimu

```bash
docker compose down
```

To also remove the database volume:

```bash
docker compose down -v
```

## Troubleshooting

### Services Not Starting

Check the logs:

```bash
docker compose logs -f
```

### Database Connection Issues

Ensure PostgreSQL is healthy:

```bash
docker compose ps
```

The `postgres` service should show `healthy` status.

### OIDC Login Failing

1. Ensure Dex is running: `docker compose logs dex`
2. Check the backend logs: `docker compose logs backend`
3. Verify the `OIDC_PROVIDERS` environment variable is set correctly

## Next Steps

- [Installation Guide](/getting-started/installation/) - Production deployment
- [Authentication](/configuration/authentication/) - Configure real OIDC providers
- [Environment Variables](/configuration/environment-variables/) - Full configuration reference
