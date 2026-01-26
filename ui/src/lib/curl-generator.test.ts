/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { generateCurlCommand } from './curl-generator';
import { Span } from '@/types/trace';

describe('generateCurlCommand', () => {
  it('generates a curl command for a tool span', () => {
    const span: Span = {
      id: '1',
      name: 'get_weather',
      type: 'tool',
      startTime: 0,
      endTime: 100,
      status: 'success',
      input: { city: 'London' }
    };

    const curl = generateCurlCommand(span);
    // In jsdom environment, window.location.origin is http://localhost:3000
    expect(curl).toContain('curl -X POST http://localhost:3000/mcp');
    expect(curl).toContain('"method":"tools/call"');
    expect(curl).toContain('"name":"get_weather"');
    expect(curl).toContain('"city":"London"');
  });

  it('generates a curl command for a service span', () => {
    const span: Span = {
      id: '2',
      name: 'fetch_data',
      type: 'service',
      startTime: 0,
      endTime: 100,
      status: 'success',
      input: {
        url: 'https://api.example.com/data',
        method: 'POST',
        headers: { 'Authorization': 'Bearer 123' },
        body: { foo: 'bar' }
      }
    };

    const curl = generateCurlCommand(span);
    expect(curl).toContain('curl -X POST "https://api.example.com/data"');
    expect(curl).toContain('-H "Authorization: Bearer 123"');
    expect(curl).toContain('-d \'{"foo":"bar"}\'');
  });

  it('handles service span without url', () => {
      const span: Span = {
        id: '3',
        name: 'fetch_data',
        type: 'service',
        startTime: 0,
        endTime: 100,
        status: 'success',
        input: {}
      };
      const curl = generateCurlCommand(span);
      expect(curl).toContain('# Unable to generate curl');
  });
});
