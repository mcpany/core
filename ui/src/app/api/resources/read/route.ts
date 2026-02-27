/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const uri = searchParams.get("uri");

  if (!uri) {
    return NextResponse.json({ error: "Missing 'uri' parameter" }, { status: 400 });
  }

  // Mock server-side response if backend not available
  // In real app, this proxies to backend
  // For test, we might want a simple echo or pass-through
  // But wait, the client.ts calls fetchWithAuth('/api/v1/resources/read?uri=...')
  // That path is intercepted by next.config.ts rewrites OR middleware.
  // The error "Failed to proxy http://localhost:50050/api/v1/resources [AggregateError: ] { code: 'ECONNREFUSED' }"
  // suggests the backend is down during test.

  // Since we are mocking in Playwright using `page.route`, the browser requests should be intercepted.
  // The client code runs in browser.
  // Why did it time out?
  // Playwright `page.route` intercepts browser network requests.
  // `apiClient.readResource` calls `fetchWithAuth`.
  // `fetchWithAuth` calls `fetch`.
  // `fetch` should be intercepted by `page.route`.

  // Ah, `fetchWithAuth` might be using absolute URL?
  // getBaseUrl() checks window.location.origin.

  return NextResponse.json({ error: "Not implemented in mock" }, { status: 501 });
}
