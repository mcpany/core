/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './test-data';

// CUJ 18-19: Profile & Collection Management
test.describe('MCP Any Profile & Collection Tests', () => {

  test.beforeEach(async ({ page, request }) => {
    // Clean start
    await request.post('/api/v1/debug/seed', {
        data: {
            upstream_services: [],
            credentials: [],
            secrets: [],
            profiles: [],
            users: []
        },
        headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
    });

    // Seed user to avoid race conditions with other tests cleanup
    await seedUser(request, 'profile-admin-e2e');

    await page.goto('/login');
    await page.fill('input[name="username"]', 'profile-admin-e2e');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
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

    // Save
    await page.getByRole('button', { name: /save|create/i }).click();

    // Verify it's created and appears
    await expect(page.getByRole('dialog')).toBeHidden();
    await expect(page.getByText('QA Profile')).toBeVisible();
  });

  test('Create Collection', async ({ page }) => {
    await page.goto('/settings/collections');
    // Implement similar flow...
  });
});
