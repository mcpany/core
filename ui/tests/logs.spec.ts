/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Log Stream Page', () => {
  test('should display log stream page', async ({ page }) => {
    // Navigate to the logs page
    await page.goto('/logs');

    // Check for the main heading
    await expect(page.getByText('Live Stream')).toBeVisible();

    // Check for control buttons
    await expect(page.getByRole('button', { name: 'Pause' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Clear' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Export' })).toBeVisible();
  });

  test('should receive and display logs', async ({ page }) => {
    await page.goto('/logs');

    // Wait for at least one log entry to appear
    // The mock backend sends a log every 800ms
    await page.waitForSelector('[data-testid="log-rows-container"] > div', { timeout: 10000 });

    const logEntries = await page.locator('[data-testid="log-rows-container"] > div').count();
    expect(logEntries).toBeGreaterThan(0);
  });

  test('should filter logs', async ({ page }) => {
    await page.goto('/logs');

    // Wait for logs
    await page.waitForSelector('[data-testid="log-rows-container"] > div');

    // Type a random string that likely won't exist to ensure filtering works (clearing the view)
    // Or better, wait for a specific common log text if predictable.
    // Since logs are random, let's type "SearchTermThatDoesntExist" and expect empty

    await page.getByPlaceholder('Filter logs...').fill('SearchTermThatDoesntExist');

    // Should be empty or show "Waiting for logs..." placeholder if cleared
    // The component shows "Waiting for logs..." if empty array
    await expect(page.getByText('Waiting for logs...')).toBeVisible();

    // Now clear filter
    await page.getByPlaceholder('Filter logs...').fill('');

    // Should see logs again
    await expect(page.getByText('Waiting for logs...')).not.toBeVisible();
    await expect(page.locator('[data-testid="log-rows-container"] > div').first()).toBeVisible();
  });

  test('should pause stream', async ({ page }) => {
    await page.goto('/logs');
    await page.waitForSelector('[data-testid="log-rows-container"] > div');

    // Click pause
    await page.getByRole('button', { name: 'Pause' }).click();

    // Verify button text changes to Resume
    await expect(page.getByRole('button', { name: 'Resume' })).toBeVisible();

    // Wait a bit to ensure potential inflight logs are rendered or ignored (if logic was perfect)
    // But since `isPaused` just stops new events from being *added* to state,
    // there might be a race where one event was processed right before pause state updated.
    await page.waitForTimeout(1000);

    // Get current log count
    const count1 = await page.locator('[data-testid="log-rows-container"] > div').count();

    // Wait a bit more to ensure no NEW logs are added
    await page.waitForTimeout(3000);

    // Get count again
    const count2 = await page.locator('[data-testid="log-rows-container"] > div').count();

    // It's possible 1 log slips through if logic isn't perfectly synchronous with UI event,
    // but usually React state update is fast.
    // If this fails, it means the Pause logic in the component isn't effective immediately
    // or the test is flaky due to timing.
    expect(count2).toBe(count1);

    // Resume
    await page.getByRole('button', { name: 'Resume' }).click();

    // Wait for new logs
    // We need to wait longer than the interval (800ms)
    await page.waitForTimeout(3000);

    const count3 = await page.locator('[data-testid="log-rows-container"] > div').count();
    expect(count3).toBeGreaterThan(count2);
  });
});
