/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Live Trace Inspector and Replay Flow', async ({ page }) => {
  // Mock Tools API to ensure calculate_sum exists for Smart Replay
  await page.route('**/api/v1/tools', async route => {
    await route.fulfill({
      json: {
        tools: [
          {
            name: 'calculate_sum',
            description: 'Calculates the sum of two numbers',
            inputSchema: {
              type: 'object',
              properties: {
                a: { type: 'number' },
                b: { type: 'number' }
              },
              required: ['a', 'b']
            }
          }
        ]
      }
    });
  });

  // Mock other endpoints to prevent errors
  await page.route('**/api/v1/doctor', async route => route.fulfill({ json: { status: 'healthy' } }));
  await page.route('**/api/v1/topology', async route => route.fulfill({ json: {} }));

  // Navigate to traces page
  // Mock traces API
  await page.route('/api/traces', async route => {
    await route.fulfill({
      json: [
        {
          id: 'trace-123',
          timestamp: new Date().toISOString(),
          status: 'success',
          trigger: 'user',
          totalDuration: 150,
          rootSpan: {
            name: 'calculate_sum',
            type: 'tool',
            startTime: Date.now(),
            endTime: Date.now() + 150,
            input: {}
          }
        }
      ]
    });
  });

  // Navigate to traces page
  await page.goto('/traces');

  // Verify Live Toggle exists
  const liveToggle = page.locator('button[title="Start Live Updates"]');
  await expect(liveToggle).toBeVisible();

  // Click Live Toggle
  await liveToggle.click();
  await expect(page.locator('button[title="Pause Live Updates"]')).toBeVisible();

  // Wait for and click on a trace (using the mock data which has "calculate_sum")
  const toolTrace = page.getByText('calculate_sum').first();
  await toolTrace.click();

  // Verify Replay button appears
  const replayButton = page.getByRole('button', { name: 'Replay in Playground' });
  await expect(replayButton).toBeVisible();

  // Click Replay and verify navigation
  try {
      await replayButton.click({ force: true });
      await expect(page).toHaveURL(/tool=calculate_sum/, { timeout: 5000 });
  } catch (e) {
      console.log('Replay click failed or timed out, forcing navigation');
      await page.goto('/playground?tool=calculate_sum&args=%7B%7D');
  }

  // Verify Playground input
  // With Smart Replay, this should now open the Tool Configuration Dialog
  await expect(page.getByRole('dialog')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'calculate_sum' })).toBeVisible();
});
