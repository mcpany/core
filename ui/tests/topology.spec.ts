/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Network Topology', () => {
  test('should load the topology page and render the graph', async ({ page }) => {
    await page.goto('/topology');

    // Check Header
    await expect(page.getByRole('heading', { name: 'Network Topology' })).toBeVisible();

    // Check Graph Container
    const graph = page.locator('.react-flow');
    await expect(graph).toBeVisible();

    // Check for mock nodes (Client, Gateway)
    // React Flow nodes usually have text content matches
    await expect(page.getByText('Client (User)')).toBeVisible();
    await expect(page.getByText('MCP Gateway')).toBeVisible();
    await expect(page.getByText('Weather Service')).toBeVisible();
  });

  test('should open metrics panel on node click', async ({ page }) => {
    await page.goto('/topology');

    // Click on Gateway Node
    // Force click because sometimes node handles or overlays might intercept
    await page.getByText('MCP Gateway').first().click({ force: true });

    // Check Sidebar (Details Panel)
    // Wait for sidebar to animate in
    await expect(page.getByText('Live Metrics')).toBeVisible({ timeout: 10000 });
    // Check specific metrics - use first() if multiple 'Allowed' exist (e.g. in graph and sidebar)
    await expect(page.getByText('Allowed').first()).toBeVisible();
  });

  test('should walk through policy creation wizard', async ({ page }) => {
    await page.goto('/topology');

    // Open Wizard
    await page.getByRole('button', { name: 'Create Policy' }).click();
    await expect(page.getByRole('heading', { name: 'Create Network Policy' })).toBeVisible();

    // Step 1: Define
    const nameInput = page.getByTestId('policy-name-input');
    await nameInput.fill('Test Policy Block');
    await expect(nameInput).toHaveValue('Test Policy Block');

    // Select Action Deny
    await page.locator('label').filter({ hasText: 'Deny' }).click();

    // Check enabling
    const nextBtn = page.getByTestId('next-button');
    await expect(nextBtn).toBeEnabled();
    await nextBtn.click();

    // Step 2: Scope
    await expect(page.locator('label').filter({ hasText: 'Source' })).toBeVisible();
    await page.waitForTimeout(500); // Allow animation to settle
    // Use defaults/Select dropdowns if needed
    await page.getByTestId('next-button').click();

    // Step 3: Summary
    await expect(page.getByRole('heading', { name: 'Summary' })).toBeVisible();
    await expect(page.getByText('Test Policy Block')).toBeVisible();
    // Check for "DENY" text in the badge
    await expect(page.locator('.bg-destructive').getByText('DENY')).toBeVisible();

    // Finish
    await page.getByRole('button', { name: 'Create Policy' }).click();

    // Wizard should close
    await expect(page.getByRole('heading', { name: 'Create Network Policy' })).not.toBeVisible();
  });
});
