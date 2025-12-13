---
title: Installation
description: Detailed installation guide for Kaimu
---

This guide covers installing Kaimu for development and production environments.

## Development Setup

### Prerequisites

- **Go 1.21+** - Backend development
- **Bun** - Frontend package manager and runtime
- **Docker** - For PostgreSQL and Dex
- **Make** (optional) - For running common commands

### Backend Setup

1. Start PostgreSQL:

```bash
docker compose up -d postgres
```

2. Install Go dependencies:

```bash
cd backend
go mod download
```

3. Run migrations:

```bash
go run cmd/main.go migrate up
```

4. Start the backend (with hot reload):

```bash
go run cmd/main.go serve
# Or with air for hot reload:
air
```

The GraphQL API will be available at `http://localhost:3000/graphql`.

### Frontend Setup

1. Install dependencies:

```bash
cd frontend
bun install
```

2. Start the development server:

```bash
bun run dev
```

The frontend will be available at `http://localhost:4321`.

### GraphQL Code Generation

After modifying GraphQL schema files:

**Backend:**
```bash
cd backend
make gql
```

**Frontend:**
```bash
cd frontend
bun run codegen
```

## Production Deployment

### Using Docker

Build and run the production images:

```bash
# Build images
docker compose -f docker-compose.prod.yaml build

# Start services
docker compose -f docker-compose.prod.yaml up -d
```

### Environment Variables

See [Environment Variables](/configuration/environment-variables/) for the full list of configuration options.

Key production settings:

```bash
# Backend
ENV=production
JWT_SECRET=your-secure-secret-here
DBPASSWORD=your-secure-database-password

# OIDC
OIDC_PROVIDERS='[{"name":"Google","slug":"google","issuer_url":"https://accounts.google.com","client_id":"...","client_secret":"..."}]'
OIDC_BASE_URL=https://api.yourdomain.com
OIDC_FRONTEND_URL=https://yourdomain.com
```

### Database

Kaimu uses PostgreSQL. For production:

1. Use a managed PostgreSQL service (AWS RDS, Google Cloud SQL, etc.)
2. Or run PostgreSQL with proper backups and replication

Required extensions:
- `uuid-ossp` (for UUID generation)

### Reverse Proxy

Example Nginx configuration:

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # Frontend
    location / {
        proxy_pass http://localhost:4321;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # Backend API
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Next Steps

- [Environment Variables](/configuration/environment-variables/) - Configure all options
- [Authentication](/configuration/authentication/) - Set up OIDC providers
- [Self-Hosting Guide](/guides/self-hosting/) - Detailed production guide
