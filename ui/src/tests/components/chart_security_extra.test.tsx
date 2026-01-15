/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render } from '@testing-library/react';
import { ChartContainer, ChartConfig } from '../../components/ui/chart';
import { describe, it, expect, vi } from 'vitest';

vi.mock('recharts', async () => {
  const actual = await vi.importActual('recharts');
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  };
});

describe('ChartContainer CSS Injection Security', () => {
  it('should block url() in color config', () => {
    const maliciousColor = 'url(javascript:alert(1))';
    const config: ChartConfig = {
      test: { label: "Test", color: maliciousColor }
    };

    const { container } = render(
      <ChartContainer config={config}>
        <div>Chart Content</div>
      </ChartContainer>
    );

    const styleTag = container.querySelector('style');
    const content = styleTag?.innerHTML || '';

    // Since we return null if url( is found, the variable should not be set.
    expect(content).not.toContain('--color-test');
  });

  it('should sanitize dangerous characters', () => {
    const maliciousColor = 'red; } body { display: none; }';
    const config: ChartConfig = {
      test: { label: "Test", color: maliciousColor }
    };

    const { container } = render(
      <ChartContainer config={config}>
        <div>Chart Content</div>
      </ChartContainer>
    );

    const styleTag = container.querySelector('style');
    const content = styleTag?.innerHTML || '';

    // The sanitization should remove ';', '}', '{'.
    // The generated CSS will contain structural '{', '}' and ';', but not from our input.
    // The attack vector requires closing the current block with '}'.

    // We verify that the specific injection sequence is broken.
    expect(content).not.toContain('} body {');

    // We check that the content words are present but the injection syntax is gone
    expect(content).toContain('red');
    expect(content).toContain('body');
    expect(content).toContain('display: none');
  });
});
