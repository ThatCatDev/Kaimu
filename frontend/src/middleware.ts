import { defineMiddleware } from 'astro:middleware';

export const onRequest = defineMiddleware(async (_context, next) => {
  // Proxy is handled by Vercel rewrites in vercel.json
  return next();
});
