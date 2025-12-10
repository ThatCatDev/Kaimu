---
title: Setting up Google Auth
description: Configure Google as an OIDC provider for Pulse
---

This guide walks through setting up Google as an authentication provider for Pulse.

## Prerequisites

- A Google Cloud account
- Access to create OAuth credentials

## Steps

### 1. Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click the project dropdown at the top
3. Click "New Project"
4. Enter a project name (e.g., "Pulse")
5. Click "Create"

### 2. Configure OAuth Consent Screen

1. In the Cloud Console, go to **APIs & Services** > **OAuth consent screen**
2. Select **External** (or Internal if using Google Workspace)
3. Click "Create"
4. Fill in the required fields:
   - **App name**: Pulse
   - **User support email**: Your email
   - **Developer contact email**: Your email
5. Click "Save and Continue"
6. Skip scopes (defaults are fine) and click "Save and Continue"
7. Add test users if in testing mode
8. Click "Save and Continue"

### 3. Create OAuth Credentials

1. Go to **APIs & Services** > **Credentials**
2. Click "Create Credentials" > "OAuth client ID"
3. Select "Web application"
4. Configure:
   - **Name**: Pulse Web Client
   - **Authorized redirect URIs**: Add your callback URL(s)

#### Callback URLs

For local development:
```
http://localhost:3000/auth/oidc/google/callback
```

For production:
```
https://api.yourdomain.com/auth/oidc/google/callback
```

5. Click "Create"
6. Copy the **Client ID** and **Client Secret**

### 4. Configure Pulse

Add Google to your `OIDC_PROVIDERS` environment variable:

```bash
OIDC_PROVIDERS='[{
  "name": "Google",
  "slug": "google",
  "issuer_url": "https://accounts.google.com",
  "client_id": "YOUR_CLIENT_ID.apps.googleusercontent.com",
  "client_secret": "YOUR_CLIENT_SECRET",
  "scopes": "openid email profile"
}]'
```

### 5. Restart Pulse

```bash
docker compose restart backend
```

### 6. Test the Integration

1. Go to `http://localhost:4321/login`
2. Click "Continue with Google"
3. Sign in with your Google account
4. You should be redirected to the Pulse dashboard

## Multiple Environments

For multiple environments (development, staging, production), create separate OAuth clients for each:

| Environment | Callback URL |
|-------------|--------------|
| Development | `http://localhost:3000/auth/oidc/google/callback` |
| Staging | `https://api.staging.yourdomain.com/auth/oidc/google/callback` |
| Production | `https://api.yourdomain.com/auth/oidc/google/callback` |

## Publishing the App

By default, Google OAuth apps are in "Testing" mode with a 100-user limit. To remove this limit:

1. Go to **OAuth consent screen**
2. Click "Publish App"
3. Complete the verification process

For internal apps (Google Workspace only), this isn't necessary.

## Troubleshooting

### "Access blocked: This app's request is invalid"

- Verify the redirect URI exactly matches what's configured in Google Cloud Console
- Check for trailing slashes or http vs https mismatches

### "Error 400: redirect_uri_mismatch"

- The callback URL doesn't match. Update the Authorized redirect URIs in Google Cloud Console.

### "This app isn't verified"

- For testing, click "Advanced" > "Go to Pulse (unsafe)"
- For production, complete the app verification process

### User Gets Logged Out Quickly

- Check `JWT_EXPIRATION_HOURS` in your environment
- Verify cookies are being set correctly (check for HTTPS requirements)

## Security Best Practices

1. **Keep secrets secure**: Never commit `client_secret` to version control
2. **Use separate credentials**: Create different OAuth clients for each environment
3. **Restrict authorized domains**: Only add necessary redirect URIs
4. **Review permissions**: Only request scopes you need
