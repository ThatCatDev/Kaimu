import { graphql } from './client';
import type { OidcProvidersQuery, OidcProvider } from '../graphql/generated';

const OIDC_PROVIDERS_QUERY = `
  query OidcProviders {
    oidcProviders {
      slug
      name
    }
  }
`;

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

export async function getOIDCProviders(): Promise<OidcProvider[]> {
  const data = await graphql<OidcProvidersQuery>(OIDC_PROVIDERS_QUERY);
  return data.oidcProviders;
}

export function getOIDCLoginURL(providerSlug: string): string {
  const config = getRuntimeConfig();

  // If proxy mode is enabled, use the local /api/ path
  if (config.useProxy === 'true') {
    return `/api/auth/oidc/${providerSlug}/authorize`;
  }

  // In browser, use runtime config or default to localhost
  const baseUrl =
    typeof window !== 'undefined'
      ? (config.apiUrl?.replace('/graphql', '') || 'http://localhost:3000')
      : (process.env.PUBLIC_API_URL?.replace('/graphql', '') || config.apiUrl?.replace('/graphql', '') || 'http://backend:3000');

  return `${baseUrl}/auth/oidc/${providerSlug}/authorize`;
}
