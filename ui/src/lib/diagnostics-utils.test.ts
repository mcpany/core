/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { analyzeConnectionError } from './diagnostics-utils';

describe('analyzeConnectionError', () => {
  it('identifies connection refused', () => {
    const result = analyzeConnectionError('dial tcp 127.0.0.1:8080: connect: connection refused');
    expect(result.title).toBe('Connection Refused');
    expect(result.category).toBe('network');
  });

  it('identifies DNS errors', () => {
    const result = analyzeConnectionError('dial tcp: lookup non-existent-host: no such host');
    expect(result.title).toBe('Host Not Found');
    expect(result.category).toBe('configuration');
  });

  it('identifies timeouts', () => {
    const result = analyzeConnectionError('context deadline exceeded');
    expect(result.title).toBe('Connection Timeout');
    expect(result.category).toBe('network');
  });

  it('identifies auth errors (401)', () => {
    const result = analyzeConnectionError('HTTP 401 Unauthorized');
    expect(result.title).toBe('Authentication Failed');
    expect(result.category).toBe('auth');
  });

  it('identifies protocol errors (TLS)', () => {
    const result = analyzeConnectionError('remote error: tls: handshake failure');
    expect(result.title).toBe('SSL/TLS Error');
    expect(result.category).toBe('protocol');
  });

  it('handles unknown errors', () => {
    const result = analyzeConnectionError('something weird happened');
    expect(result.title).toBe('Unknown Error');
    expect(result.category).toBe('unknown');
    expect(result.description).toBe('something weird happened');
  });
});
