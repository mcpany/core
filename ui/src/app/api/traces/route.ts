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

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:50050';

export async function GET(request: Request) {
  try {
    // Forward the API Key if present in the incoming request headers
    const headers: Record<string, string> = {};
    const apiKey = request.headers.get('x-api-key') || request.headers.get('authorization');
    if (apiKey) {
        headers['Authorization'] = apiKey;
    }

    const res = await fetch(`${BACKEND_URL}/debug/entries`, {
       headers,
       cache: 'no-store'
    });

    if (!res.ok) {
        console.warn(`Backend returned ${res.status} for traces`);
        // If backend fails, fallback to empty list instead of crashing
        return NextResponse.json([]);
    }

    const entries = await res.json();
    const traces = entries.map((entry: any) => mapDebugEntryToTrace(entry));

    // Sort by timestamp descending
    traces.sort((a: Trace, b: Trace) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

    return NextResponse.json(traces);
  } catch (e) {
    console.error("Failed to fetch traces from backend", e);
    return NextResponse.json([]);
  }
}

function mapDebugEntryToTrace(entry: any): Trace {
    const startTime = new Date(entry.timestamp).getTime();
    // Duration from Go is in nanoseconds (number)
    const durationMs = entry.duration / 1e6;
    const endTime = startTime + durationMs;

    let status: SpanStatus = 'success';
    let errorMessage: string | undefined;

    if (entry.status >= 400) {
        status = 'error';
        errorMessage = `HTTP ${entry.status}`;
    }

    let input: any;
    try {
        input = entry.request_body ? JSON.parse(entry.request_body) : undefined;
    } catch {
        input = { raw: entry.request_body };
    }

    let output: any;
    try {
        output = entry.response_body ? JSON.parse(entry.response_body) : undefined;
    } catch {
        output = { raw: entry.response_body };
    }

    // Try to detect MCP method from input
    let name = `${entry.method} ${entry.path}`;
    let type: Span['type'] = 'core';
    let serviceName = 'http-gateway';

    if (input && input.method) {
        name = input.method; // e.g. "tools/call"
        if (input.params && input.params.name) {
             name = `${input.method} (${input.params.name})`;
             type = 'tool';
        }
    }

    const rootSpan: Span = {
        id: `sp_${entry.id}`,
        name: name,
        type: type,
        startTime: startTime,
        endTime: endTime,
        status: status,
        input: input,
        output: output,
        serviceName: serviceName,
        errorMessage: errorMessage,
        children: []
    };

    return {
        id: entry.id,
        rootSpan: rootSpan,
        timestamp: entry.timestamp,
        totalDuration: durationMs,
        status: status,
        trigger: 'user' // Default to user
    };
}
