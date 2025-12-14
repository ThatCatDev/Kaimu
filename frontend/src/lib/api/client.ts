// Get runtime config from window (injected by Layout.astro) or fall back to build-time env
function getRuntimeConfig() {
  if (typeof window !== 'undefined' && (window as any).__RUNTIME_CONFIG__) {
    return (window as any).__RUNTIME_CONFIG__;
  }
  return {
    apiUrl: import.meta.env.PUBLIC_API_URL || '',
    useProxy: import.meta.env.PUBLIC_USE_PROXY || '',
  };
}

function getApiUrl(): string {
  const config = getRuntimeConfig();

  // If proxy mode is enabled, use the local /api/ path (same-origin)
  if (config.useProxy === 'true') {
    return '/api/graphql';
  }

  // In browser, use runtime config or default to localhost
  if (typeof window !== 'undefined') {
    return config.apiUrl || 'http://localhost:3000/graphql';
  }
  // During SSR, use process.env or fall back to service name
  return process.env.PUBLIC_API_URL || config.apiUrl || 'http://backend:3000/graphql';
}

interface GraphQLResponse<T> {
  data?: T;
  errors?: Array<{ message: string; extensions?: { code?: string } }>;
}

interface RefreshTokenResponse {
  refreshToken: {
    success: boolean;
    expiresIn: number;
  };
}

// Track if a refresh is in progress to prevent multiple simultaneous refreshes
let refreshPromise: Promise<boolean> | null = null;
let tokenExpiresAt: number | null = null;

// Refresh token mutation query
const REFRESH_TOKEN_MUTATION = `
  mutation RefreshToken {
    refreshToken {
      success
      expiresIn
    }
  }
`;

/**
 * Attempts to refresh the access token using the refresh token cookie
 * Returns true if successful, false otherwise
 */
async function refreshAccessToken(): Promise<boolean> {
  // If a refresh is already in progress, wait for it
  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = (async () => {
    try {
      const response = await fetch(getApiUrl(), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ query: REFRESH_TOKEN_MUTATION }),
      });

      const result: GraphQLResponse<RefreshTokenResponse> = await response.json();

      if (result.errors?.length || !result.data?.refreshToken?.success) {
        return false;
      }

      // Update token expiration time
      tokenExpiresAt = Date.now() + (result.data.refreshToken.expiresIn * 1000);
      return true;
    } catch {
      return false;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

/**
 * Check if the token is expired or about to expire (within 30 seconds)
 */
function isTokenExpiredOrExpiring(): boolean {
  if (!tokenExpiresAt) {
    return false; // We don't know when it expires, let the server decide
  }
  // Refresh 30 seconds before expiration
  return Date.now() >= tokenExpiresAt - 30000;
}

/**
 * Check if an error indicates an authentication failure
 */
function isAuthError(errors?: Array<{ message: string; extensions?: { code?: string } }>): boolean {
  if (!errors?.length) return false;

  return errors.some(error => {
    const message = error.message.toLowerCase();
    return message.includes('unauthorized') ||
           message.includes('unauthenticated') ||
           message.includes('not authenticated') ||
           message.includes('session expired') ||
           message.includes('invalid token') ||
           error.extensions?.code === 'UNAUTHENTICATED';
  });
}

export async function graphql<T>(
  query: string,
  variables?: Record<string, unknown>,
  options?: { skipRefresh?: boolean }
): Promise<T> {
  // Proactively refresh if token is about to expire (only in browser)
  if (typeof window !== 'undefined' && isTokenExpiredOrExpiring() && !options?.skipRefresh) {
    await refreshAccessToken();
  }

  const response = await fetch(getApiUrl(), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({ query, variables }),
  });

  const result: GraphQLResponse<T> = await response.json();

  // If we get an auth error and we're in the browser, try to refresh and retry
  if (isAuthError(result.errors) && typeof window !== 'undefined' && !options?.skipRefresh) {
    const refreshed = await refreshAccessToken();

    if (refreshed) {
      // Retry the original request
      const retryResponse = await fetch(getApiUrl(), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ query, variables }),
      });

      const retryResult: GraphQLResponse<T> = await retryResponse.json();

      if (retryResult.errors?.length) {
        throw new Error(retryResult.errors[0].message);
      }

      if (!retryResult.data) {
        throw new Error('No data returned from GraphQL');
      }

      return retryResult.data;
    }

    // Refresh failed, throw the original error
    throw new Error(result.errors![0].message);
  }

  if (result.errors?.length) {
    throw new Error(result.errors[0].message);
  }

  if (!result.data) {
    throw new Error('No data returned from GraphQL');
  }

  return result.data;
}

/**
 * Update the token expiration time after login/register
 * Call this after successful authentication
 */
export function setTokenExpiration(expiresInSeconds: number): void {
  tokenExpiresAt = Date.now() + (expiresInSeconds * 1000);
}

/**
 * Clear the token expiration (on logout)
 */
export function clearTokenExpiration(): void {
  tokenExpiresAt = null;
}

/**
 * Manually trigger a token refresh
 * Useful for checking if the session is still valid
 */
export async function tryRefreshToken(): Promise<boolean> {
  return refreshAccessToken();
}
