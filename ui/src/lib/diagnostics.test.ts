/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { analyzeTrace } from './diagnostics';
import { Trace, Span } from '@/types/trace';

// Mock Trace Helper
function createMockTrace(status: 'success' | 'error', errorMessage?: string, outputError?: string): Trace {
  return {
    id: 'test-trace',
    timestamp: new Date().toISOString(),
    totalDuration: 100,
    status,
    trigger: 'user',
    rootSpan: {
      id: 'span-1',
      name: 'test-tool',
      type: 'tool',
      startTime: Date.now(),
      endTime: Date.now() + 100,
      status,
      errorMessage,
      output: outputError ? { error: outputError } : {},
      children: []
    }
  };
}

describe('analyzeTrace', () => {
  it('should return no diagnostics for success trace', () => {
    const trace = createMockTrace('success');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(0);
  });

  it('should detect schema validation errors', () => {
    const trace = createMockTrace('error', 'Schema validation error: property "foo" missing');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('Schema Validation Error');
    expect(diagnostics[0].type).toBe('error');
  });

  it('should detect Zod errors', () => {
    const trace = createMockTrace('error', 'ZodError: invalid type');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('Schema Validation Error');
  });

  it('should detect permission errors', () => {
    const trace = createMockTrace('error', undefined, 'EPERM: operation not permitted');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('Permission Denied');
  });

  it('should detect JSON parse errors', () => {
    const trace = createMockTrace('error', 'SyntaxError: Unexpected token < in JSON at position 0');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('JSON Parsing Error');
  });

  it('should detect timeout errors', () => {
    const trace = createMockTrace('error', 'Deadline exceeded');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('Operation Timed Out');
  });

  it('should detect connection errors', () => {
    const trace = createMockTrace('error', 'connect: connection refused');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('Connection Failed');
  });

  it('should detect authentication errors', () => {
    const trace = createMockTrace('error', 'HTTP 401 Unauthorized');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('Authentication Failed');
  });

  it('should detect missing tool errors', () => {
    const trace = createMockTrace('error', 'Tool not found: my-tool');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('Tool Not Found');
  });

  it('should detect rate limit errors', () => {
    const trace = createMockTrace('error', '429 Too Many Requests');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('Rate Limit Exceeded');
  });

  it('should detect recursion depth', () => {
    const trace = createMockTrace('error', 'some error');
    // Create deep chain
    let current = trace.rootSpan;
    for (let i = 0; i < 11; i++) {
        const child: Span = {
            id: `span-${i}`,
            name: 'recursive-span',
            type: 'tool',
            startTime: Date.now(),
            endTime: Date.now(),
            status: 'success',
            children: []
        };
        current.children = [child];
        current = child;
    }

    const diagnostics = analyzeTrace(trace);
    // Should have Unknown Error AND Recursion Warning
    expect(diagnostics).toHaveLength(2);
    expect(diagnostics.find(d => d.title === 'High Recursion Depth')).toBeDefined();
  });

  it('should provide fallback for unknown errors', () => {
    const trace = createMockTrace('error', 'Something went wrong');
    const diagnostics = analyzeTrace(trace);
    expect(diagnostics).toHaveLength(1);
    expect(diagnostics[0].title).toBe('Unknown Error');
    expect(diagnostics[0].message).toBe('Something went wrong');
  });
});
