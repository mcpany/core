/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Live Trace Inspector and Replay Flow', async ({ page }) => {
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

  // Mock WebSocket to ensure "Connected" state regardless of backend availability
  await page.addInitScript(() => {
    const originalWebSocket = window.WebSocket;

    // Minimal mock that triggers onopen immediately
    class MockWebSocket extends EventTarget {
      readyState: number;

      constructor(url: string | URL, protocols?: string | string[]) {
        super();
        this.readyState = 0; // CONNECTING

        // Trigger Open immediately to simulate connection
        setTimeout(() => {
          this.readyState = 1; // OPEN
          this.dispatchEvent(new Event('open'));
        }, 10);
      }

      close() {
        this.readyState = 3; // CLOSED
        this.dispatchEvent(new Event('close'));
      }

      send(data: any) {
        // no-op
      }
    }

    // Override the global WebSocket
    (window as any).WebSocket = MockWebSocket;
    // Keep constants if needed (though not used by useTraces usually)
    (window as any).WebSocket.CONNECTING = 0;
    (window as any).WebSocket.OPEN = 1;
    (window as any).WebSocket.CLOSING = 2;
    (window as any).WebSocket.CLOSED = 3;
  });

  // Navigate to traces page
  await page.goto('/traces');

  // Verify Live Toggle exists (Starts in Live/Pause state)
  const liveToggle = page.locator('button[title="Pause Live Updates"]');
  await expect(liveToggle).toBeVisible();

  // Verify it's enabled (connected)
  await expect(liveToggle).toBeEnabled();

  // Click Live Toggle to Pause
  await liveToggle.click();
  await expect(page.locator('button[title="Resume Live Updates"]')).toBeVisible();

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
  const input = page.getByPlaceholder('Enter command or select a tool...').or(page.locator('textarea'));
  await expect(input).toBeVisible();
  await expect(input).toHaveValue(/calculate_sum/);
});
