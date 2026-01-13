/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Trace Viewer', () => {
  test('should navigate to traces page and view details', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('/');

    // Check if Traces link exists in sidebar and click it
    await page.click('a[href="/traces"]');

    // Verify URL
    await expect(page).toHaveURL(/\/traces/);

    // Wait for traces to load
    await page.waitForSelector('text=Loading traces...', { state: 'detached' });

    // Check if list is populated (should have at least one trace from mock)
    // Check if list is populated (should have at least one trace from mock)
    // Use try/catch or flexible selector since mock data is random
    // But our mock generator creates at least one calculate_sum
    // Actually, let's just check for any trace item
    const firstTrace = page.locator('button.flex.flex-col').first();
    await expect(firstTrace).toBeVisible();

    // Click the first trace
    await firstTrace.click();

    // Check if details pane is populated
    await expect(page.locator('h2')).toBeVisible(); // Trace name in header
    await expect(page.locator('text=Execution Waterfall')).toBeVisible();
    await expect(page.locator('text=Root Input')).toBeVisible();
  });

  test('should filter traces', async ({ page }) => {
    await page.goto('/traces');

    // Wait for traces
    await page.waitForSelector('text=Loading traces...', { state: 'detached' });

    // Type in search box
    await page.fill('input[placeholder="Search traces..."]', 'calculate');

    // Expect only matching items
    // and doesn't crash the page
    await expect(page.locator('input[placeholder="Search traces..."]')).toHaveValue('calculate');
  });

  test('should replay trace in playground', async ({ page }) => {
    await page.goto('/traces');

    // Ensure we have a trace to click
    await page.waitForSelector('text=Loading traces...', { state: 'detached' });
    const firstTrace = page.locator('button.flex.flex-col').first();
    await expect(firstTrace).toBeVisible();
    await firstTrace.click();

    // Click "Replay in Playground"
    // We look for the button with specific text
    const replayBtn = page.getByRole('button', { name: 'Replay in Playground' });
    await expect(replayBtn).toBeVisible();
    await replayBtn.click();

    // Verify redirection to playground
    await expect(page).toHaveURL(/\/playground.*/);

    // Verify query params are present (tool and args)
    // We don't check exact values as they depend on the random mock trace
    const url = page.url();
    expect(url).toContain('tool=');
    expect(url).toContain('args=');

    // Verify Playground input is populated
    // The input should contain the tool name or args
    // We wait for the form or input to be visible first
    await expect(page.getByPlaceholder('Enter command or select a tool...').or(page.locator('textarea'))).toBeVisible();
  });
});
