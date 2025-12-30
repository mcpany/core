/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import React from 'react';
import { render } from '@testing-library/react';
import { ChartContainer, ChartConfig } from '../../components/ui/chart';
import { describe, it, expect } from 'vitest';

// Mock RechartsResponsiveContainer to avoid issues with Recharts in JSDOM
// We need to mock the module 'recharts'
import * as Recharts from 'recharts';

// Mocking Recharts components
vi.mock('recharts', async () => {
  const actual = await vi.importActual('recharts');
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  };
});


describe('ChartContainer Security', () => {
  it('should not allow escaping CSS context in color config', () => {
    const maliciousConfig: ChartConfig = {
        malicious: {
            label: "Malicious",
            color: "red; } body { display: none; } .foo { --color-bar: blue",
        }
    };

    const { container } = render(
      <ChartContainer config={maliciousConfig}>
        <div>Chart Content</div>
      </ChartContainer>
    );

    const styleTag = container.querySelector('style');
    expect(styleTag).toBeInTheDocument();

    // In a vulnerable version, the malicious string is injected directly.
    // We want to verify if the raw string is present.
    // The injected CSS would look like: --color-malicious: red; } body { display: none; } .foo { --color-bar: blue;
    const content = styleTag?.innerHTML || '';

    // Ideally, we want to ensure that specific dangerous characters are handled.
    // For this test, we just check if the raw malicious string is present.
    expect(content).not.toContain('body { display: none; }');
  });

  it('should not allow XSS via closing style tag', () => {
     const maliciousConfig: ChartConfig = {
        xss: {
            label: "XSS",
            color: "red; </style><script>alert(1)</script><style>",
        }
    };

    const { container } = render(
      <ChartContainer config={maliciousConfig}>
        <div>Chart Content</div>
      </ChartContainer>
    );

    const styleTag = container.querySelector('style');
    const content = styleTag?.innerHTML || '';

    // If vulnerable, this will contain the closing style tag
    expect(content).not.toContain('</style>');
  });
});
