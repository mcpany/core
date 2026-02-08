/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextRequest, NextResponse } from 'next/server';

/**
 * Handles GET requests to download resource content.
 *
 * @param request - The incoming NextRequest.
 * @returns A NextResponse containing the resource content as a download, or an error.
 */
export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const uri = searchParams.get('uri');
  const name = searchParams.get('name');
  const token = searchParams.get('token');

  if (!uri || !name) {
    return NextResponse.json({ error: 'Missing uri or name' }, { status: 400 });
  }

  const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';
  const headers: HeadersInit = {};

  // Pass auth token from client if provided
  if (token) {
    headers['Authorization'] = `Basic ${token}`;
  } else if (process.env.MCPANY_API_KEY) {
    // Fallback to server-side API Key if configured
    headers['X-API-Key'] = process.env.MCPANY_API_KEY;
  }

  try {
    const res = await fetch(`${backendUrl}/api/v1/resources/read?uri=${encodeURIComponent(uri)}`, {
      headers,
      cache: 'no-store'
    });

    if (!res.ok) {
      console.error(`Failed to fetch resource: ${res.status} ${res.statusText}`);
      return NextResponse.json({ error: 'Failed to fetch resource' }, { status: res.status });
    }

    const data = await res.json();
    if (!data.contents || data.contents.length === 0) {
      return NextResponse.json({ error: 'Resource content not found' }, { status: 404 });
    }

    const content = data.contents[0];
    let body: BodyInit;

    if (content.blob) {
      // Decode base64 using Buffer
      body = Buffer.from(content.blob, 'base64');
    } else {
      body = content.text || '';
    }

    const response = new NextResponse(body);
    response.headers.set('Content-Type', content.mimeType || 'application/octet-stream');
    response.headers.set('Content-Disposition', `attachment; filename="${name}"`);

    return response;

  } catch (error) {
    console.error('Error in download route:', error);
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
  }
}
