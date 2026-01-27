/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Traces page loads and shows sequence diagram', async ({ page }) => {
  // Mock API response
  await page.route('*/**/api/traces', async route => {
    const json = [
      {
        id: 'trace-1',
        rootSpan: {
          id: 'span-1',
          name: 'test_tool',
          type: 'tool',
          startTime: 1672531200000,
          endTime: 1672531200100,
          status: 'success',
          input: { foo: 'bar' },
          output: { result: 'ok' },
          children: []
        },
        timestamp: '2023-01-01T00:00:00.000Z',
        totalDuration: 100,
        status: 'success',
        trigger: 'user'
      }
    ];
    await route.fulfill({ json });
  });

  await page.goto('/traces');

  // Wait for trace list item and click it
  await page.getByText('test_tool').first().click();

  // Check if tabs are present
  await expect(page.getByRole('tab', { name: 'Sequence Diagram' })).toBeVisible();

  // Click Sequence Diagram tab
  await page.getByRole('tab', { name: 'Sequence Diagram' }).click();

  // Check for error first
  const errorMessage = page.getByText('Failed to render sequence diagram.');
  if (await errorMessage.isVisible()) {
      throw new Error('Mermaid rendering failed');
  }

  // Verify diagram renders
  // We check for the SVG element within the container
  const container = page.getByTestId('mermaid-container');
  await expect(container).toBeVisible();
  await expect(container.locator('svg')).toBeVisible({ timeout: 10000 });

  // Verify content inside diagram
  await expect(container.getByText('test_tool').first()).toBeVisible();
});
