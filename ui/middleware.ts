import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const pathname = request.nextUrl.pathname;

  // Intercept /api/v1 requests AND gRPC requests
  if (pathname.startsWith('/api/v1') || pathname.startsWith('/mcpany.api.v1.')) {
    const requestHeaders = new Headers(request.headers);

    // Inject API Key from server-side environment variable
    const apiKey = process.env.MCPANY_API_KEY;
    if (apiKey) {
      requestHeaders.set('X-API-Key', apiKey);
    }

    // Pass the modified headers to the next step (which allows rewrites to pick them up)
    return NextResponse.next({
      request: {
        headers: requestHeaders,
      },
    });
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/api/v1/:path*',
    '/mcpany.api.v1.:path*',
  ],
};
