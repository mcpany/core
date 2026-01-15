/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Logs Page', () => {

  // Mock WebSocket for stable testing
  test.beforeEach(async ({ page }) => {
    // Mock the WebSocket connection
    await page.routeWebSocket(/\/api\/ws\/logs/, ws => {
        // Simply accept the connection mock (no action needed, just don't connect to server)
        // We can save it if needed, but per-test override is better
    });

    await page.goto('/logs');
  });

  test('should display logs title', async ({ page }) => {
    await expect(page).toHaveTitle(/MCPAny/);
    await expect(page.getByRole('heading', { name: 'Live Logs' })).toBeVisible();
  });

  test('should display log entries', async ({ page }) => {
    // Send a mock log
    await page.routeWebSocket('**/api/v1/ws/logs', ws => {
        ws.send(JSON.stringify({
            id: '1',
            timestamp: new Date().toISOString(),
            level: 'INFO',
            message: 'Test Log Entry',
            source: 'test'
        }));
    });

    await page.reload(); // Reload to trigger new connection with mock

    const logs = page.getByTestId('log-rows-container').locator('.group');
    // Filter to find our specific log, ignoring potential background noise
    await expect(logs.filter({ hasText: 'Test Log Entry' })).toBeVisible();
  });

  test.skip('should pause on scroll up', async ({ page }) => {
    // Send enough logs to overflow
    await page.routeWebSocket('**/api/v1/ws/logs', ws => {
        for (let i = 0; i < 50; i++) {
            ws.send(JSON.stringify({
                id: `${i}`,
                timestamp: new Date().toISOString(),
                level: 'INFO',
                message: `Log Entry ${i} - filling space to enable scrolling`,
                source: 'test'
            }));
        }
    });

    await page.reload();

    const logContainer = page.getByTestId('log-rows-container');
    const logs = page.getByTestId('log-rows-container').locator('.group');
    await expect(logs.filter({ hasText: 'Log Entry' })).toHaveCount(50);

    // Hover over container to ensure focus
    await logContainer.hover();

    // Scroll up manually via evaluate to ensure event fires
    await page.evaluate(() => {
        const viewport = document.querySelector('[data-radix-scroll-area-viewport]');
        if (viewport) {
            viewport.scrollTop = 0;
            viewport.dispatchEvent(new Event('scroll'));
        }
    });

    // Check for paused state (Resume button)
    await expect(page.getByRole('button', { name: 'Resume' })).toBeVisible();
  });

  test('should pause and resume logs', async ({ page }) => {
     let socketRoute: any;

     await page.routeWebSocket('**/api/v1/ws/logs', ws => {
        socketRoute = ws;
        // Initial log
        ws.send(JSON.stringify({
            id: 'init',
            timestamp: new Date().toISOString(),
            level: 'INFO',
            message: 'Initial Log',
            source: 'test'
        }));
    });

    await page.reload();
    const logs = page.getByTestId('log-rows-container').locator('.group');
    await expect(logs.filter({ hasText: 'Initial Log' })).toBeVisible();

    // Click pause
    const pauseButton = page.getByRole('button', { name: 'Pause' });
    await pauseButton.click();
    await expect(page.getByRole('button', { name: 'Resume' })).toBeVisible();

    // Send another log while paused
    if (socketRoute) {
        socketRoute.send(JSON.stringify({
            id: 'paused',
            timestamp: new Date().toISOString(),
            level: 'INFO',
            message: 'Igored Log',
            source: 'test'
        }));
    }

    // Wait short time to ensure UI process it (and ignores it)
    await page.waitForTimeout(1000);
    // Should NOT see Ignored Log (checking count doesn't increase for THAT log)
    // But we can just check filtering
    await expect(logs.filter({ hasText: 'Igored Log' })).toBeHidden();

    // Resume
    const resumeButton = page.getByRole('button', { name: 'Resume' });
    await resumeButton.click();
    await expect(page.getByRole('button', { name: 'Pause' })).toBeVisible();

    // Send new log
    if (socketRoute) {
         socketRoute.send(JSON.stringify({
            id: 'resumed',
            timestamp: new Date().toISOString(),
            level: 'INFO',
            message: 'New Log',
            source: 'test'
        }));
    }

    await expect(logs.filter({ hasText: 'New Log' })).toBeVisible();
  });

  test('should filter logs', async ({ page }) => {
    await page.routeWebSocket('**/api/v1/ws/logs', ws => {
        ws.send(JSON.stringify({ id: '1', timestamp: new Date().toISOString(), level: 'INFO', message: 'Info Log', source: 'test' }));
        ws.send(JSON.stringify({ id: '2', timestamp: new Date().toISOString(), level: 'ERROR', message: 'Error Log', source: 'test' }));
    });

    await page.reload();
    const logs = page.getByTestId('log-rows-container').locator('.group');
    // Wait for both
    await expect(logs.filter({ hasText: 'Info Log' })).toBeVisible();
    await expect(logs.filter({ hasText: 'Error Log' })).toBeVisible();


    // Filter by ERROR
    const filterSelect = page.getByRole('combobox');
    await filterSelect.click();
    await page.getByRole('option', { name: 'Error' }).click();

    await expect(logs).toHaveCount(1);
    await expect(logs.first()).toContainText('Error Log');
    await expect(logs.first()).not.toContainText('Info Log');
  });

  test('should clear logs', async ({ page }) => {
    await page.routeWebSocket('**/api/v1/ws/logs', ws => {
        ws.send(JSON.stringify({ id: '1', timestamp: new Date().toISOString(), level: 'INFO', message: 'Log 1', source: 'test' }));
        ws.send(JSON.stringify({ id: '2', timestamp: new Date().toISOString(), level: 'INFO', message: 'Log 2', source: 'test' }));
    });

    await page.reload();
    const logs = page.getByTestId('log-rows-container').locator('.group');
    await expect(logs).toHaveCount(2);

    const clearButton = page.getByRole('button', { name: 'Clear' });
    await clearButton.click();

    await expect(logs).toHaveCount(0);
  });
});
