/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render } from '@testing-library/react';
import { ChartContainer, ChartConfig } from '../../components/ui/chart';
import { describe, it, expect, vi } from 'vitest';

// Mock Recharts
vi.mock('recharts', async () => {
  const actual = await vi.importActual('recharts');
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  };
});

describe('ChartContainer Security', () => {
  it('should not allow XSS via id prop escaping style tag', () => {
    const maliciousId = 'test</style><div id="injected">XSS</div><style>';
    const config: ChartConfig = {
      test: { label: "Test", color: "red" }
    };

    const { container } = render(
      <ChartContainer id={maliciousId} config={config}>
        <div>Chart Content</div>
      </ChartContainer>
    );

    const styleTag = container.querySelector('style');
    const content = styleTag?.innerHTML || '';

    // The ID is sanitized by replacing non-alphanumeric chars with -
    expect(content).not.toContain('</style>');
    // It should contain the sanitized ID
    // The regex /[^a-zA-Z0-9\-_]/g replaces special chars with -
    // test</style><div id="injected">XSS</div><style> -> test--style--div-id--injected--XSS--div--style-
    expect(content).toContain('test--style--div-id--injected--XSS--div--style-');
  });

  it('should sanitize CSS values in config', () => {
    const maliciousConfig: ChartConfig = {
      test: {
        label: "Test",
        // Try to break out of the property: "red; background: url(javascript:alert(1))"
        // Try to close style tag: "red</style><script>alert(1)</script>"
        color: "red; background: url(javascript:alert(1)); </style><script>alert(1)</script>"
      }
    };

    const { container } = render(
      <ChartContainer config={maliciousConfig}>
        <div>Chart Content</div>
      </ChartContainer>
    );

    const styleTag = container.querySelector('style');
    const content = styleTag?.innerHTML || '';

    // Verify allowed characters only
    // Colon, semicolon, slash, angle brackets should be stripped.
    // "red; background: url(javascript:alert(1)); </style><script>alert(1)</script>"
    // -> "red background url(javascriptalert(1)) style scriptalert(1)script"
    // Wait, regex is /[^a-zA-Z0-9\-_#%.(),\s]/g
    // ; -> stripped
    // : -> stripped
    // / -> stripped
    // < -> stripped
    // > -> stripped

    expect(content).not.toContain(';');
    expect(content).not.toContain(':');
    expect(content).not.toContain('<');
    expect(content).not.toContain('>');
    expect(content).not.toContain('url('); // url( is blocked by specific check too
  });
});
