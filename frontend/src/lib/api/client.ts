function getApiUrl(): string {
  // In browser, always use localhost (browser can't reach docker internal network)
  if (typeof window !== 'undefined') {
    return import.meta.env.PUBLIC_API_URL || 'http://localhost:3000/graphql';
  }
  // During SSR in Docker, use the service name
  return import.meta.env.PUBLIC_API_URL || 'http://backend:3000/graphql';
}

interface GraphQLResponse<T> {
  data?: T;
  errors?: Array<{ message: string }>;
}

export async function graphql<T>(
  query: string,
  variables?: Record<string, unknown>
): Promise<T> {
  const response = await fetch(getApiUrl(), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({ query, variables }),
  });

  const result: GraphQLResponse<T> = await response.json();

  if (result.errors?.length) {
    throw new Error(result.errors[0].message);
  }

  if (!result.data) {
    throw new Error('No data returned from GraphQL');
  }

  return result.data;
}
