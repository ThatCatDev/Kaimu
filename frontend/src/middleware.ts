import { defineMiddleware } from 'astro:middleware';

export const onRequest = defineMiddleware(async (context, next) => {
  // Read at runtime for Vercel serverless compatibility
  const PROXY_BACKEND_URL = process.env.PROXY_BACKEND_URL || import.meta.env.PROXY_BACKEND_URL || '';

  // Debug logging
  console.log('[Middleware] Path:', context.url.pathname);
  console.log('[Middleware] PROXY_BACKEND_URL:', PROXY_BACKEND_URL ? 'SET' : 'NOT SET');

  // Only proxy if PROXY_BACKEND_URL is set and path starts with /api/
  if (PROXY_BACKEND_URL && context.url.pathname.startsWith('/api/')) {
    console.log('[Middleware] Proxying to:', PROXY_BACKEND_URL);
    const backendPath = context.url.pathname.replace(/^\/api/, '');
    const backendUrl = `${PROXY_BACKEND_URL}${backendPath}${context.url.search}`;

    try {
      // Forward the request to the backend
      const headers = new Headers(context.request.headers);
      // Remove host header to avoid issues
      headers.delete('host');

      const response = await fetch(backendUrl, {
        method: context.request.method,
        headers,
        body: context.request.method !== 'GET' && context.request.method !== 'HEAD'
          ? await context.request.text()
          : undefined,
        // Forward credentials
        credentials: 'include',
      });

      // Create response with backend's headers
      const responseHeaders = new Headers(response.headers);

      // Forward cookies from backend
      const setCookie = response.headers.get('set-cookie');
      if (setCookie) {
        responseHeaders.set('set-cookie', setCookie);
      }

      return new Response(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers: responseHeaders,
      });
    } catch (error) {
      console.error('Proxy error:', error);
      return new Response(JSON.stringify({ error: 'Proxy error' }), {
        status: 502,
        headers: { 'Content-Type': 'application/json' },
      });
    }
  }

  return next();
});
