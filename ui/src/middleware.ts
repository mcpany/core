/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const pathname = request.nextUrl.pathname;
  const requestHeaders = new Headers(request.headers);

  // Intercept /api/v1 requests AND gRPC requests
  if (pathname.startsWith('/api/v1') || pathname.startsWith('/mcpany.api.v1.') || pathname.startsWith('/doctor') || pathname.startsWith('/v1/') || pathname.startsWith('/auth/oauth/') || pathname === '/auth/login' || pathname.startsWith('/debug/')) {
    // Inject API Key from server-side environment variable
    const apiKey = process.env.MCPANY_API_KEY;
    if (apiKey) {
      requestHeaders.set('X-API-Key', apiKey);
    }

    // Dynamic Proxying via Middleware to avoid build-time baking of BACKEND_URL
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:50059';
    console.log(`[Middleware] Proxying ${pathname} to ${backendUrl}`);

    const url = new URL(request.url);
    const newUrl = new URL(pathname + url.search, backendUrl);

    return NextResponse.rewrite(newUrl, {
      request: {
        headers: requestHeaders,
      },
    });
  }

  const response = NextResponse.next({
    request: {
      headers: requestHeaders,
    },
  });

  // Add Security Headers
  const csp = [
    "default-src 'self'",
    "script-src 'self' 'unsafe-eval' 'unsafe-inline' https://cdn.jsdelivr.net", // Added cdn.jsdelivr.net for Monaco Editor
    "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdn.jsdelivr.net", // Added cdn.jsdelivr.net
    "img-src 'self' data: https:",
    "font-src 'self' data: https://fonts.gstatic.com",
    "connect-src 'self' https://cdn.jsdelivr.net", // Added cdn.jsdelivr.net
    "worker-src 'self' blob:", // Added for Monaco Editor workers
    "frame-ancestors 'none'",
    "object-src 'none'",
    "base-uri 'self'"
  ].join('; ');

  response.headers.set('Content-Security-Policy', csp);
  response.headers.set('X-Frame-Options', 'DENY');
  response.headers.set('X-Content-Type-Options', 'nosniff');
  response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
  response.headers.set('Permissions-Policy', 'camera=(), microphone=(), geolocation=()');
  response.headers.set('Strict-Transport-Security', 'max-age=63072000; includeSubDomains; preload');

  return response;
}

export const config = {
  matcher: [
    // Apply to all routes except static files
    '/((?!_next/static|_next/image|favicon.ico).*)',
  ],
};
