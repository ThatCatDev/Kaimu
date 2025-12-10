---
title: Environment Variables
description: Complete reference for Pulse environment variables
---

This page documents all environment variables used to configure Pulse.

## Backend Configuration

### Application

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Port for the GraphQL API server |
| `ENV` | `development` | Environment (`development` or `production`) |
| `JWT_SECRET` | `dev-secret-change-in-production` | Secret key for signing JWT tokens. **Change in production!** |
| `JWT_EXPIRATION_HOURS` | `24` | JWT token expiration time in hours |

### Database

| Variable | Default | Description |
|----------|---------|-------------|
| `DBHOST` | `localhost` | PostgreSQL host |
| `DBNAME` | `pulse` | Database name |
| `DBUSERNAME` | `pulse` | Database user |
| `DBPASSWORD` | *required* | Database password |
| `DBPORT` | `5432` | Database port |
| `DBSSL` | `disable` | SSL mode (`disable`, `require`, `verify-full`) |

### OIDC Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `OIDC_BASE_URL` | `http://localhost:3000` | Backend URL (for OIDC callbacks) |
| `OIDC_FRONTEND_URL` | `http://localhost:4321` | Frontend URL (for redirects after auth) |
| `OIDC_STATE_EXPIRATION_MINUTES` | `10` | OIDC state expiration time |
| `OIDC_PROVIDERS` | *empty* | JSON array of OIDC provider configurations |

## Frontend Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PUBLIC_API_URL` | `http://localhost:3000/graphql` | GraphQL API endpoint URL |

## OIDC Providers Configuration

The `OIDC_PROVIDERS` environment variable accepts a JSON array of provider configurations:

```bash
OIDC_PROVIDERS='[
  {
    "name": "Google",
    "slug": "google",
    "issuer_url": "https://accounts.google.com",
    "client_id": "your-client-id.apps.googleusercontent.com",
    "client_secret": "your-client-secret",
    "scopes": "openid email profile"
  }
]'
```

### Provider Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Display name shown on login button |
| `slug` | Yes | URL-safe identifier (used in `/auth/oidc/{slug}/callback`) |
| `issuer_url` | Yes | OIDC issuer URL |
| `discovery_url` | No | Alternative URL for OIDC discovery (for Docker networking) |
| `client_id` | Yes | OAuth client ID |
| `client_secret` | Yes | OAuth client secret |
| `scopes` | No | Space-separated scopes (default: `openid email profile`) |

### Multiple Providers

```bash
OIDC_PROVIDERS='[
  {"name":"Google","slug":"google","issuer_url":"https://accounts.google.com","client_id":"...","client_secret":"..."},
  {"name":"Okta","slug":"okta","issuer_url":"https://dev-123456.okta.com","client_id":"...","client_secret":"..."},
  {"name":"Azure AD","slug":"azure","issuer_url":"https://login.microsoftonline.com/TENANT_ID/v2.0","client_id":"...","client_secret":"..."}
]'
```

## Docker Compose Example

```yaml
services:
  backend:
    environment:
      # Database
      DBHOST: postgres
      DBNAME: pulse
      DBUSERNAME: pulse
      DBPASSWORD: ${DB_PASSWORD}
      DBPORT: 5432
      DBSSL: disable

      # Application
      PORT: 3000
      ENV: production
      JWT_SECRET: ${JWT_SECRET}
      JWT_EXPIRATION_HOURS: 24

      # OIDC
      OIDC_BASE_URL: https://api.yourdomain.com
      OIDC_FRONTEND_URL: https://yourdomain.com
      OIDC_STATE_EXPIRATION_MINUTES: 10
      OIDC_PROVIDERS: ${OIDC_PROVIDERS}

  frontend:
    environment:
      PUBLIC_API_URL: https://api.yourdomain.com/graphql
```

## Using a .env File

Create a `.env` file in the project root:

```bash
# Database
DB_PASSWORD=your-secure-password

# JWT
JWT_SECRET=your-secure-jwt-secret

# OIDC Providers
OIDC_PROVIDERS='[{"name":"Google","slug":"google","issuer_url":"https://accounts.google.com","client_id":"...","client_secret":"..."}]'
```

Docker Compose will automatically load this file.

## Security Notes

1. **Never commit secrets** - Use environment variables or secret management
2. **Change JWT_SECRET** - The default is insecure
3. **Use HTTPS in production** - Required for secure cookies
4. **Rotate secrets regularly** - Especially client secrets
