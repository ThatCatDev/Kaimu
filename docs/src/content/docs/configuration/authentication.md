---
title: Authentication (OIDC)
description: Configure OpenID Connect authentication for Kaimu
---

Kaimu uses OpenID Connect (OIDC) for authentication. This allows you to integrate with any OIDC-compliant identity provider.

## How It Works

1. User clicks a provider button on the login page
2. Browser redirects to the identity provider
3. User authenticates with the provider
4. Provider redirects back to Kaimu with an authorization code
5. Backend exchanges code for tokens and creates/links user
6. User is logged into Kaimu

## Configuration

OIDC providers are configured via the `OIDC_PROVIDERS` environment variable as a JSON array:

```bash
OIDC_PROVIDERS='[
  {
    "name": "Provider Name",
    "slug": "provider-slug",
    "issuer_url": "https://provider.example.com",
    "client_id": "your-client-id",
    "client_secret": "your-client-secret",
    "scopes": "openid email profile"
  }
]'
```

## Provider Configuration Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Display name shown on the login button (e.g., "Google", "Okta") |
| `slug` | Yes | URL-safe identifier used in callback URLs. Must be unique. |
| `issuer_url` | Yes | OIDC issuer URL from your provider |
| `discovery_url` | No | Alternative URL for OIDC discovery (for Docker/internal networks) |
| `client_id` | Yes | OAuth 2.0 client ID from your provider |
| `client_secret` | Yes | OAuth 2.0 client secret from your provider |
| `scopes` | No | Space-separated OAuth scopes (default: `openid email profile`) |

## Callback URL

When configuring your OIDC provider, you'll need to register a callback/redirect URL:

```
{OIDC_BASE_URL}/auth/oidc/{slug}/callback
```

For example:
- Local development: `http://localhost:3000/auth/oidc/google/callback`
- Production: `https://api.yourdomain.com/auth/oidc/google/callback`

## Supported Providers

Kaimu works with any OIDC-compliant provider. Here are some common ones:

### Google

- Issuer URL: `https://accounts.google.com`
- [Setup Guide](/guides/google-auth/)

### Okta

- Issuer URL: `https://dev-XXXXXX.okta.com` (or your custom domain)
- [Setup Guide](/guides/okta-auth/)

### Azure AD / Microsoft Entra ID

- Issuer URL: `https://login.microsoftonline.com/{tenant-id}/v2.0`

### Auth0

- Issuer URL: `https://{your-tenant}.auth0.com/`

### Keycloak

- Issuer URL: `https://{keycloak-host}/realms/{realm}`

### Dex (Local Development)

- Issuer URL: `http://localhost:5556/dex`
- Discovery URL: `http://dex:5556/dex` (for Docker networking)

## Multiple Providers

You can configure multiple providers. Users will see all enabled providers on the login page:

```bash
OIDC_PROVIDERS='[
  {"name":"Google","slug":"google","issuer_url":"https://accounts.google.com","client_id":"...","client_secret":"..."},
  {"name":"Microsoft","slug":"microsoft","issuer_url":"https://login.microsoftonline.com/common/v2.0","client_id":"...","client_secret":"..."},
  {"name":"Company SSO","slug":"company","issuer_url":"https://sso.company.com","client_id":"...","client_secret":"..."}
]'
```

## User Linking

When a user authenticates via OIDC:

1. **Existing OIDC Identity**: If the user has logged in with this provider before, they're logged into their existing account.

2. **Email Match**: If a user with the same verified email exists, the OIDC identity is linked to that account.

3. **New User**: If no match is found, a new user account is created.

This allows users to:
- Log in with multiple providers linked to the same account
- Migrate from username/password to OIDC
- Use different providers on different devices

## Docker Networking (discovery_url)

When running in Docker, the backend may not be able to reach `localhost` URLs. Use `discovery_url` to specify an internal URL for OIDC discovery:

```bash
OIDC_PROVIDERS='[{
  "name": "Dex",
  "slug": "dex",
  "issuer_url": "http://localhost:5556/dex",      # Browser sees this
  "discovery_url": "http://dex:5556/dex",         # Backend uses this
  "client_id": "kaimu-app",
  "client_secret": "kaimu-secret-key"
}]'
```

## JWT Token Management

Kaimu uses a dual-token authentication system for security:

### Token Types

| Token | Lifetime | Storage | Purpose |
|-------|----------|---------|---------|
| Access Token | 5 minutes | HTTP-only cookie | Short-lived token for API requests |
| Refresh Token | 7 days | HTTP-only cookie + database | Long-lived token for obtaining new access tokens |

### How It Works

1. **Login/OIDC Callback**: User authenticates and receives both tokens as HTTP-only cookies
2. **API Requests**: Access token is automatically sent with each request
3. **Token Refresh**: When the access token expires (or 30 seconds before), the frontend automatically uses the refresh token to get a new pair
4. **Token Rotation**: Each refresh generates a new refresh token and invalidates the old one
5. **Logout**: Both tokens are revoked and cookies are cleared

### Security Features

- **Short Access Token Lifetime**: Limits exposure if a token is compromised
- **Token Rotation**: Refresh tokens are single-use; a new one is issued on each refresh
- **Reuse Detection**: If a refresh token is used after rotation (indicating theft), all user tokens are revoked
- **HTTP-only Cookies**: Tokens cannot be accessed by JavaScript, preventing XSS attacks
- **Secure Cookies**: In production, cookies require HTTPS
- **Device Tracking**: Refresh tokens store user agent and IP for auditing

### Frontend Auto-Refresh

The frontend automatically handles token refresh:
- Proactively refreshes tokens 30 seconds before expiration
- Retries failed requests after successful refresh
- Redirects to login on authentication failure

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | - | Secret key for signing tokens. **Required in production** |
| `JWT_ACCESS_EXPIRATION_MINUTES` | `5` | Access token lifetime |
| `JWT_REFRESH_EXPIRATION_DAYS` | `7` | Refresh token lifetime |

## Security Considerations

1. **Use HTTPS in Production**: OIDC requires secure connections for token exchange. Cookies are only secure over HTTPS.

2. **Protect Client Secrets**: Never commit secrets to version control. Use environment variables or secret management.

3. **Verify Issuer URLs**: Ensure issuer URLs match exactly what your provider specifies.

4. **Review Scopes**: Only request scopes you need. The default `openid email profile` is sufficient for most cases.

5. **State Expiration**: The default 10-minute state expiration helps prevent CSRF attacks.

6. **Change JWT_SECRET**: The default secret is insecure. Use a strong, random secret in production.

## Troubleshooting

### "Provider not found"

- Check that the `slug` in the URL matches a configured provider
- Verify `OIDC_PROVIDERS` is valid JSON

### "Invalid issuer"

- Ensure `issuer_url` matches exactly what the provider returns
- Check for trailing slashes

### "Failed to get OIDC provider metadata"

- Verify the provider is reachable from the backend
- Check `discovery_url` for Docker environments
- Look at backend logs for detailed errors

### "Token exchange failed"

- Verify `client_id` and `client_secret` are correct
- Check that the callback URL is registered with the provider
- Ensure the authorization code hasn't expired
