/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Summary: Represents the status of a span.
 *
 * Params:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export type SpanStatus = 'success' | 'error' | 'pending';

/**
 * Summary: Represents a span in a trace.
 *
 * Params:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
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
 * Summary: Represents a full trace.
 *
 * Params:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export interface Trace {
  id: string;
  rootSpan: Span;
  timestamp: string;
  totalDuration: number;
  status: SpanStatus;
  trigger: 'user' | 'webhook' | 'scheduler' | 'system';
}
