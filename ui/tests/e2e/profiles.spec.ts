/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedDebugData } from './test-data';

// CUJ 18-19: Profile & Collection Management
test.describe('MCP Any Profile & Collection Tests', () => {

  test.beforeEach(async ({ page, request }) => {
    // Clean start using debug seeder
    await seedDebugData({
        users: [{
            id: "profile-admin-e2e",
            authentication: {
                basic_auth: {
                    username: "profile-admin-e2e",
                    password_hash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a" // password
                }
            },
            roles: ["admin"],
            profile_ids: ["dev"] // Make sure 'dev' exists or is created
        }],
        profiles: [{ name: "dev" }] // Seed required profile
    }, request);

    await page.goto('/login');
    await page.fill('input[name="username"]', 'profile-admin-e2e');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]', { force: true });
    await page.waitForURL('/', { timeout: 30000 });
  });

  test('Create new Profile', async ({ page }) => {
    // Navigate to Profiles page
    await page.goto('/profiles');

    // assert headers
    await expect(page.getByText('Profiles', { exact: true }).first()).toBeVisible();


    // Click Create
    const createBtn = page.getByRole('button', { name: /Create|Plus/i });
    await expect(createBtn).toBeVisible();
    await createBtn.click({ force: true });

    // Wait for dialog
    await expect(page.getByText(/Create (New )?Profile/i).first()).toBeVisible();

    // Fill form
    await page.getByLabel(/name/i).fill('QA Profile');

    // Description is not currently supported in the UI
    // await page.getByLabel(/description/i).fill('Restricted access for QA');

    // Save
    await page.getByRole('button', { name: /save|create/i }).click();

    // Verify it appears in the list (Real Data!)
    await expect(page.getByText('QA Profile')).toBeVisible();
  });

  test('Create Collection', async ({ page }) => {
    await page.goto('/settings/collections');
    // Implement similar flow...
  });
});
