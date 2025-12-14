import type { APIRoute } from 'astro';

const handler: APIRoute = async ({ params, request }) => {
  const PROXY_BACKEND_URL = process.env.PROXY_BACKEND_URL || import.meta.env.PROXY_BACKEND_URL || '';

  console.log('[API Route] Path:', params.path);
  console.log('[API Route] PROXY_BACKEND_URL:', PROXY_BACKEND_URL ? 'SET' : 'NOT SET');

  if (!PROXY_BACKEND_URL) {
    return new Response(JSON.stringify({ error: 'Proxy not configured' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  const backendPath = params.path ? `/${params.path}` : '';
  const url = new URL(request.url);
  const backendUrl = `${PROXY_BACKEND_URL}${backendPath}${url.search}`;

  console.log('[API Route] Proxying to:', backendUrl);

  try {
    const headers = new Headers(request.headers);
    headers.delete('host');

    const response = await fetch(backendUrl, {
      method: request.method,
      headers,
      body: request.method !== 'GET' && request.method !== 'HEAD'
        ? await request.text()
        : undefined,
    });

    const responseHeaders = new Headers(response.headers);

    // Remove compression headers - fetch already decompresses
    responseHeaders.delete('content-encoding');
    responseHeaders.delete('content-length');

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
    console.error('[API Route] Proxy error:', error);
    return new Response(JSON.stringify({ error: 'Proxy error' }), {
      status: 502,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};

// Export handler for all HTTP methods
export const GET = handler;
export const POST = handler;
export const PUT = handler;
export const DELETE = handler;
export const PATCH = handler;
export const OPTIONS = handler;
