---
title: Setting up Okta
description: Configure Okta as an OIDC provider for Kaimu
---

This guide walks through setting up Okta as an authentication provider for Kaimu.

## Prerequisites

- An Okta account (free developer account works)
- Admin access to your Okta organization

## Steps

### 1. Create an Okta Application

1. Log in to your [Okta Admin Console](https://admin.okta.com/)
2. Go to **Applications** > **Applications**
3. Click "Create App Integration"
4. Select:
   - **Sign-in method**: OIDC - OpenID Connect
   - **Application type**: Web Application
5. Click "Next"

### 2. Configure the Application

Fill in the application settings:

- **App integration name**: Kaimu
- **Grant type**: Authorization Code (default)
- **Sign-in redirect URIs**:
  - Development: `http://localhost:3000/auth/oidc/okta/callback`
  - Production: `https://api.yourdomain.com/auth/oidc/okta/callback`
- **Sign-out redirect URIs**: (optional)
  - `http://localhost:4321/login`
  - `https://yourdomain.com/login`
- **Controlled access**: Choose your assignment option

Click "Save"

### 3. Get Client Credentials

After creating the app:

1. Note the **Client ID** on the application page
2. Click "Edit" in the Client Credentials section
3. Note the **Client Secret** (or generate one if needed)

### 4. Find Your Issuer URL

Your Okta issuer URL is:

```
https://{your-okta-domain}/oauth2/default
```

Or if using a custom authorization server:
```
https://{your-okta-domain}/oauth2/{authorization-server-id}
```

You can find this at **Security** > **API** > **Authorization Servers**

### 5. Configure Kaimu

Add Okta to your `OIDC_PROVIDERS` environment variable:

```bash
OIDC_PROVIDERS='[{
  "name": "Okta",
  "slug": "okta",
  "issuer_url": "https://dev-XXXXXX.okta.com/oauth2/default",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "scopes": "openid email profile"
}]'
```

### 6. Restart Kaimu

```bash
docker compose restart backend
```

### 7. Test the Integration

1. Go to `http://localhost:4321/login`
2. Click "Continue with Okta"
3. Sign in with your Okta account
4. You should be redirected to the Kaimu dashboard

## Assigning Users

### Individual Assignment

1. Go to **Applications** > **Kaimu**
2. Click the "Assignments" tab
3. Click "Assign" > "Assign to People"
4. Select users and click "Assign"

### Group Assignment

1. Go to **Applications** > **Kaimu**
2. Click the "Assignments" tab
3. Click "Assign" > "Assign to Groups"
4. Select groups and click "Assign"

## Custom Authorization Server

For production, consider creating a custom authorization server:

1. Go to **Security** > **API**
2. Click "Add Authorization Server"
3. Configure:
   - **Name**: Kaimu API
   - **Audience**: `kaimu`
   - **Description**: Authorization server for Kaimu
4. Use the custom issuer URL in your Kaimu configuration

## Troubleshooting

### "Invalid redirect_uri"

- Verify the redirect URI exactly matches what's configured in Okta
- Check for trailing slashes or protocol mismatches

### "User is not assigned to the client application"

- Assign the user to the Kaimu application in Okta
- Or enable "Allow everyone to self-service" in the application settings

### "Invalid scope"

- Verify the authorization server has the required scopes enabled
- Go to **Security** > **API** > **Authorization Servers** > **Scopes**

### CORS Errors

Okta handles CORS automatically for registered redirect URIs. If you see CORS errors:
- Verify your frontend URL is correct
- Check that the backend is handling the callback, not the frontend

## Security Best Practices

1. **Use a custom authorization server** for production workloads
2. **Enable MFA** for Okta users
3. **Review session policies** for appropriate timeout settings
4. **Monitor sign-in activity** in the Okta System Log
5. **Use groups** to manage access at scale
