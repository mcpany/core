/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page, expect } from '@playwright/test';

/**
 * Logs in to the application using the UI.
 * Assumes the user "e2e-admin" with password "password" exists (seeded via test-data.ts).
 */
export async function login(page: Page) {
    await page.goto('/login');
    // Wait for page to be fully loaded as it might be transitioning
    await page.waitForLoadState('networkidle');

    await page.fill('input[name="username"]', 'e2e-admin');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');

    // Wait for redirect to home page and verify
    await expect(page).toHaveURL('/', { timeout: 15000 });
}
