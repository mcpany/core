/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Represents the status of a span.
 */
export type SpanStatus = 'success' | 'error' | 'pending';

/**
 * Represents a span in a trace.
 */
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

/**
 * Represents a full trace.
 */
export interface Trace {
  id: string;
  rootSpan: Span;
  timestamp: string;
  totalDuration: number;
  status: SpanStatus;
  trigger: 'user' | 'webhook' | 'scheduler' | 'system';
}
