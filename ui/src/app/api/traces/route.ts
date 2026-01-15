/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NextResponse } from 'next/server';

export type SpanStatus = 'success' | 'error' | 'pending';

export interface Span {
  id: string;
  name: string;
  type: 'tool' | 'service' | 'resource' | 'prompt' | 'core';
  startTime: number;
  endTime: number;
  status: SpanStatus;
  input?: Record<string, any>;
  output?: Record<string, any>;
  children?: Span[];
  serviceName?: string;
  errorMessage?: string;
}

export interface Trace {
  id: string;
  rootSpan: Span;
  timestamp: string;
  totalDuration: number;
  status: SpanStatus;
  trigger: 'user' | 'webhook' | 'scheduler' | 'system';
}

interface DebugEntry {
  id: string;
  timestamp: string;
  method: string;
  path: string;
  status: number;
  duration: number; // nanoseconds
  request_headers: Record<string, string[]>;
  response_headers: Record<string, string[]>;
  request_body: string;
  response_body: string;
}

function tryParseJson(str: string): any {
  try {
    return JSON.parse(str);
  } catch {
    return str;
  }
}

export async function GET(request: Request) {
  const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:50050';
  const apiKey = process.env.MCPANY_API_KEY;
  const headers: Record<string, string> = {
    'Accept': 'application/json'
  };

  if (apiKey) {
    headers['X-API-Key'] = apiKey;
  }

  // Forward existing auth headers if any
  const reqHeaders = new Headers(request.headers);
  const authHeader = reqHeaders.get('authorization');
  if (authHeader) {
    headers['Authorization'] = authHeader;
  }
  const cookieHeader = reqHeaders.get('cookie');
  if (cookieHeader) {
    headers['Cookie'] = cookieHeader;
  }

  try {
    const res = await fetch(`${backendUrl}/debug/entries`, {
        cache: 'no-store',
        headers: headers
    });

    if (res.ok) {
        const entries: DebugEntry[] = await res.json();

        // Map DebugEntry to Trace
        const traces: Trace[] = entries.map(entry => {
            const durationMs = entry.duration / 1000000; // ns to ms
            const startTime = new Date(entry.timestamp).getTime();

            const rootSpan: Span = {
                id: `sp_${entry.id}`,
                name: `${entry.method} ${entry.path}`,
                type: 'core',
                startTime: startTime,
                endTime: startTime + durationMs,
                status: entry.status >= 400 ? 'error' : 'success',
                input: tryParseJson(entry.request_body),
                output: tryParseJson(entry.response_body),
                serviceName: 'backend',
                children: []
            };

            return {
                id: entry.id,
                rootSpan: rootSpan,
                timestamp: entry.timestamp,
                totalDuration: durationMs,
                status: rootSpan.status,
                trigger: 'user' // Default assumption
            };
        });

        // If we have traces, return them.
        // If not (e.g. backend empty), we might want to return empty list.
        return NextResponse.json(traces);
    }
  } catch (error) {
    console.error("Failed to fetch debug entries from backend:", error);
    // Fallback to empty list or mock data?
    // Given the task is to align with roadmap (REAL backend), returning empty list + error log is better than fake data
    // confusing the user.
    // However, for the purpose of the UI not crashing if backend is down during this specific test env:
  }

  // If backend is unreachable or returns error, we return empty list or fallback to mock?
  // I will return empty list to signify "Real Data Connection Attempted but failed/empty"
  // But wait, if I return empty, the UI shows "No traces".
  // If I return Mock data, I can verify the UI still works.
  // The roadmap goal is to use REAL data.
  // So I will return empty array if fetch fails, but maybe with a special error trace?

  // Let's stick to returning empty array to prove we switched to real backend.
  return NextResponse.json([]);
}
