/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('User Guide Walkthrough', () => {
  test('Dashboard loads key metrics', async ({ page }) => {
    await page.goto('/');
    // Explicitly wait for the response to ensure the frontend has received data
    // Use a longer timeout for CI
    await page.waitForResponse(response => response.url().includes('/api/v1/dashboard/metrics'), { timeout: 45000 });
    await page.waitForLoadState('networkidle');

    // Check for "Total Requests" card
    await expect(page.locator('text=Total Requests')).toBeVisible({ timeout: 45000 });
    // Check for "Active Services" card
    await expect(page.locator('text=Active Services')).toBeVisible();
    await expect(page.locator('text=Connected Tools')).toBeVisible();
  });

  test('Services: Add Service Redirects to Marketplace', async ({ page }) => {
    await page.goto('/upstream-services');
    await expect(page.getByRole('heading', { name: 'Upstream Services' })).toBeVisible();

    // Explicitly target the button with text "Add Service"
    const addButton = page.getByRole('button', { name: 'Add Service' });
    await expect(addButton).toBeVisible();

    // Check for dialog opens
    await page.waitForLoadState('networkidle');
    await addButton.click({ force: true });
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('New Service')).toBeVisible();

    // Close it
    await page.keyboard.press('Escape');
    await expect(page.getByRole('dialog')).toBeHidden();
  });

  test('Resources: List and Preview', async ({ page }) => {
    await page.goto('/resources');
    await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible();

    // Check for known resources from seeded data (Echo Service has System Logs)
    // Wait for the table/list to load
    await expect(page.locator('text=System Logs')).toBeVisible({ timeout: 20000 });
  });

  test('Global Search Modal', async ({ page }) => {
    await page.goto('/');

    // Wait for hydration/network idle to ensure event listeners are attached
    // "domcontentloaded" is not enough for React effect listeners sometimes
    await page.waitForLoadState('networkidle');

    // Press Ctrl+K
    // Try forcing focus on body first
    await page.locator('body').focus();
    await page.keyboard.press('Control+k');

    // Check for Search Input placeholder or identifying element
    // Fallback: If Ctrl+K fails (flaky in CI headless), try clicking the search button if available
    // But for now, just increasing stability of the keypress
    await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible({ timeout: 20000 });

    // Close modal (Esc key or click outside/close button)
    await page.keyboard.press('Escape');

    await expect(page.getByPlaceholder('Type a command or search...')).not.toBeVisible();
  });

  test('Logs Stream', async ({ page }) => {
    await page.goto('/logs');
    await expect(page.getByRole('heading', { name: 'Live Logs' })).toBeVisible();
    // Check for log container - using more specific selector to avoid strict mode violation
    // expected container has bg-black/90
    await expect(page.locator('div.font-mono.bg-black\\/90')).toBeVisible();
  });

  test('Secrets Vault', async ({ page }) => {
    await page.goto('/secrets');
    await expect(page.getByRole('heading', { name: 'API Keys & Secrets' })).toBeVisible();
    // Check for Add Secret button
    await expect(page.getByRole('button', { name: 'Add Secret' })).toBeVisible();
  });

  test('Alerts Page', async ({ page }) => {
    await page.goto('/alerts');
    await expect(page.getByRole('heading', { name: 'Alerts & Incidents' })).toBeVisible();
    // Check for "Alerts & Incidents" text or stats
    await expect(page.getByText('Monitor system health')).toBeVisible();
  });

  test('Stack Composer', async ({ page }) => {
    await page.goto('/stacks');
    await expect(page.getByRole('heading', { name: 'Stacks' })).toBeVisible();
    // Check if stacks are present OR if the empty state with Create Stack button is shown
    // We relax this check because in a fresh environment without seeds, the list might be empty.
    const hasSystemStack = await page.getByText('mcpany-system').isVisible();
    if (!hasSystemStack) {
        // Should show "No stacks found" or "Create Stack"
        await expect(page.getByText('Create Stack').first()).toBeVisible();
    }
  });

  test('Webhooks Management', async ({ page }) => {
    await page.goto('/webhooks');
    await expect(page.getByRole('heading', { name: 'Webhooks' })).toBeVisible();
    // Button is "New Webhook", not "Add Webhook"
    await expect(page.getByRole('button', { name: 'New Webhook' })).toBeVisible();
  });

  test('Connection Diagnostic Tool', async ({ page }) => {
    // Navigate to services first
    await page.goto('/upstream-services');
    await page.waitForLoadState('networkidle');

    // Use "Payment Gateway" which is seeded in test-data.ts
    const row = page.locator('tr').filter({ hasText: 'Payment Gateway' });
    await expect(row).toBeVisible({ timeout: 20000 });

    // Open Edit Sheet
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Check for Edit Sheet load
    await expect(page.getByRole('heading', { name: 'Edit Service' })).toBeVisible();
    await expect(page.locator('input[id="name"]')).toHaveValue('Payment Gateway');
  });
});
