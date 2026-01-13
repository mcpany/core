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

describe('ChartContainer ID XSS Security', () => {
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

    // If vulnerable, the style tag will be closed and the div will be rendered as HTML
    // However, React renders `id` prop safely in the <div>.
    // The vulnerability is in `ChartStyle` which uses `dangerouslySetInnerHTML`.

    const styleTag = container.querySelector('style');
    const content = styleTag?.innerHTML || '';

    // If vulnerable, content will contain the closing style tag literally
    // AND the subsequent injected HTML might be interpretted by browser (but JSDOM might just show text in style).
    // The key is that `dangerouslySetInnerHTML` puts the string raw into the style element.
    // If the string contains `</style>`, the browser will close the style block.

    // We check if the malicious string is present in the style tag content.
    expect(content).not.toContain('</style><div id="injected">XSS</div>');
  });
});
