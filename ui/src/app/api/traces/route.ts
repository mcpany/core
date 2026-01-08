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

// Helper to generate mock spans
function generateSpan(
  id: string,
  name: string,
  type: Span['type'],
  startOffset: number,
  duration: number,
  serviceName?: string
): Span {
  const now = Date.now();
  // We'll base timestamps relative to a fixed start for consistency in the response,
  // but "now" works for fresh data.
  // Actually, let's just use offsets for the mock generator logic
  return {
    id,
    name,
    type,
    startTime: startOffset,
    endTime: startOffset + duration,
    status: Math.random() > 0.9 ? 'error' : 'success', // 10% chance of error
    serviceName,
    input: { query: "example input", param: 123 },
    output: { result: "example output", confidence: 0.99 },
    children: []
  };
}

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  // const limit = searchParams.get('limit');

  const traces: Trace[] = [];
  const now = Date.now();

  // Generate 20 mock traces
  for (let i = 0; i < 20; i++) {
    const traceId = `tr_${Math.random().toString(36).substring(2, 9)}`;
    const startTime = now - (i * 1000 * 60 * 5); // spaced out by 5 mins

    // Scenario 1: Simple Calculator
    if (i % 3 === 0) {
      const root = generateSpan(`sp_${traceId}_1`, "calculate_sum", "tool", startTime, 150, "math-service");
      root.status = 'success'; // force success for root
      root.input = { a: 10, b: 20 };
      root.output = { result: 30 };

      traces.push({
        id: traceId,
        rootSpan: root,
        timestamp: new Date(startTime).toISOString(),
        totalDuration: 150,
        status: 'success',
        trigger: 'user'
      });
    }
    // Scenario 2: RAG Chain (Complex)
    else if (i % 3 === 1) {
      const totalDuration = 2500;
      const root = generateSpan(`sp_${traceId}_1`, "research_topic", "core", startTime, totalDuration, "orchestrator");

      // Step 1: Search
      const searchSpan = generateSpan(`sp_${traceId}_2`, "google_search", "tool", startTime + 50, 800, "search-service");
      searchSpan.input = { query: "MCP architecture patterns" };
      searchSpan.output = { hits: 5, top_url: "..." };
      root.children!.push(searchSpan);

      // Step 2: Scrape (parallel-ish)
      const scrapeSpan = generateSpan(`sp_${traceId}_3`, "web_scraper", "tool", startTime + 900, 1200, "browser-service");
      root.children!.push(scrapeSpan);

      // Step 3: Summarize
      const summarizeSpan = generateSpan(`sp_${traceId}_4`, "llm_summarize", "tool", startTime + 2150, 300, "llm-gateway");
      root.children!.push(summarizeSpan);

      const isError = Math.random() > 0.8;
      if (isError) {
        scrapeSpan.status = 'error';
        scrapeSpan.errorMessage = "Timeout waiting for selector .content";
        root.status = 'error';
      } else {
        root.status = 'success';
      }

      traces.push({
        id: traceId,
        rootSpan: root,
        timestamp: new Date(startTime).toISOString(),
        totalDuration,
        status: root.status,
        trigger: 'webhook'
      });
    }
    // Scenario 3: Database Query
    else {
      const duration = 45;
      const root = generateSpan(`sp_${traceId}_1`, "get_user_profile", "resource", startTime, duration, "user-db");
      root.status = 'success';

      traces.push({
        id: traceId,
        rootSpan: root,
        timestamp: new Date(startTime).toISOString(),
        totalDuration: duration,
        status: 'success',
        trigger: 'system'
      });
    }
  }

  return NextResponse.json(traces);
}
