/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Global Search', () => {
  test.beforeEach(async ({ page, request }) => {
    // Seed data for search
    // Seed data for search
    try {
        const r = await request.post('/api/v1/services', {
            data: {
               id: "svc_02",
               name: "User Service",
               disable: false,
               version: "v1.0",
               grpc_service: { address: "localhost:50051", tools: [], resources: [] }
            }
        });
        expect(r.ok()).toBeTruthy();
    } catch (e) {
        console.error("Seeding failed", e);
        throw e;
    }
    await page.goto('/');
  });

  test('should open command palette by clicking the search button', async ({ page }) => {
     // Find the button with text "Search" or similar
     // The button text changes based on screen size ("Search feature..." vs "Search...")
     const searchButton = page.locator('button').filter({ hasText: /Search/ }).first();
     await searchButton.click();

     const commandDialog = page.locator('[role="dialog"]');
     await expect(commandDialog).toBeVisible();
  });

  test('should open command palette via shortcut and search dynamic content', async ({ page }) => {
    // Focus page to ensure keyboard events are captured
    await page.click('body');

    // Determine modifier key
    const modifier = process.platform === 'darwin' ? 'Meta' : 'Control';
    await page.keyboard.press(`${modifier}+k`);
    await page.waitForTimeout(500);

    const dialog = page.locator('div[role="dialog"]');
    await expect(dialog).toBeVisible({ timeout: 10000 });
    const searchInput = page.locator('input[placeholder*="Type a command or search"]');
    await expect(searchInput).toBeVisible();

    // Check availability of content types
    await searchInput.fill('User Service'); // Filtering explicitly
    await expect(page.getByRole('option', { name: /User Service/i }).first()).toBeVisible();

    // Navigate to it
    await page.keyboard.press('Enter');
    // We expect navigation or action. For now, checking the search worked is finding the item.
  });
});
