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

export async function getOIDCProviders(): Promise<OidcProvider[]> {
  const data = await graphql<OidcProvidersQuery>(OIDC_PROVIDERS_QUERY);
  return data.oidcProviders;
}

export function getOIDCLoginURL(providerSlug: string): string {
  // In browser, always use localhost (browser can't reach docker internal network)
  const baseUrl =
    typeof window !== 'undefined'
      ? import.meta.env.PUBLIC_API_URL?.replace('/graphql', '') ||
        'http://localhost:3000'
      : import.meta.env.PUBLIC_API_URL?.replace('/graphql', '') ||
        'http://backend:3000';

  return `${baseUrl}/auth/oidc/${providerSlug}/authorize`;
}
