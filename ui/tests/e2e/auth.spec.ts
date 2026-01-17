/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

// Use a unique ID to avoid collisions
const TEST_TIMESTAMP = Date.now();
const USER_ID = `e2e-user-${TEST_TIMESTAMP}`;

test.describe('Authentication and User Management', () => {
  // We can't easily test login without enabling it in the backend config first,
  // effectively locking out the runner unless we configure it.
  // For now, we assume the environment might be running in "no-auth" or "basic-auth" mode.
  // But we can test the UI functionality of the User Management page.

  test.beforeEach(async ({ page }) => {
    // Mock List Users (initially empty or some defaults)
    await page.route('**/api/v1/users', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({ json: { users: [] } });
        } else if (route.request().method() === 'POST') {
             // Mock Create
             // We can dynamically update the list for subsequent GETs if we wanted,
             // but strictly for this test we can just return success and mock the NEXT get.
             const postData = route.request().postDataJSON();
             await route.fulfill({ json: { user: postData.user } });
        } else {
            await route.continue();
        }
    });

    await page.goto('/');
  });

  test.skip('should render login page components', async ({ page }) => {
    await page.goto('/login');

    // Check for essential elements mentioned in auth.md
    await expect(page.getByRole('heading', { name: /Sign In|Login/i })).toBeVisible();
    await expect(page.getByLabel(/Email|Username/i)).toBeVisible();
    await expect(page.getByLabel(/Password/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /Sign In|Login/i })).toBeVisible();
  });

  test('should navigate to user management', async ({ page }) => {
    // Open sidebar if closed? It's open by default on desktop.
    await page.goto('/users');
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();
  });

  test('should create, edit, and delete a user', async ({ page }) => {
    await page.goto('/users');

    // Create User
    await page.getByRole('button', { name: 'Add User' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Create User' })).toBeVisible();

    await page.getByLabel('Username').fill(USER_ID);
    await page.getByLabel('Role').fill('viewer');
    await page.getByLabel('Password').fill('password123'); // Optional for basic auth

    // Update mock for the next GET request to include the new user
    // We need to override the route handler.
    // Since Playwright routes are handle first-match or specific order, we can unroute or just handle it smarter in beforeEach.
    // Simpler: Just override it *before* the action that triggers the re-fetch (which happens on Dialog close/submit).
    // The previous POST route handler was generic.
    // Let's refine the test structure to just mock the "After Create" state.

    await page.route('**/api/v1/users', async route => {
        if (route.request().method() === 'GET') {
             await route.fulfill({ json: { users: [{ id: USER_ID, roles: ['viewer'], authentication: { basic_auth: {} } }] } });
        } else {
             await route.fulfill({ json: {} }); // POST/PUT success
        }
    });

    await page.getByRole('button', { name: 'Save changes' }).click();

    // Verify user in list
    // Use .first() if multiple matches, but unique ID helps
    await expect(page.getByRole('cell', { name: USER_ID })).toBeVisible();
    await expect(page.getByText('viewer', { exact: true })).toBeVisible();

    // Edit User
    // Find the row with the user, then click edit
    const row = page.getByRole('row', { name: USER_ID });

    // Mock the List update for AFTER edit
    await page.route('**/api/v1/users', async route => {
        if (route.request().method() === 'GET') {
             await route.fulfill({ json: { users: [{ id: USER_ID, roles: ['editor'], authentication: { basic_auth: {} } }] } });
        }
    });
    // Mock the PUT request
    await page.route(`**/api/v1/users/${USER_ID}`, async route => {
         await route.fulfill({ json: {} });
    });

    await row.getByRole('button').first().click(); // First button is Edit (Pencil)

    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Edit User' })).toBeVisible();

    // Check if fields are populated
    await expect(page.getByLabel('Username')).toBeDisabled(); // ID is immutable usually
    await expect(page.getByLabel('Role')).toHaveValue('viewer');

    await page.getByLabel('Role').fill('editor');
    await page.getByRole('button', { name: 'Save changes' }).click();

    // Verify update
    await expect(page.getByText('editor', { exact: true })).toBeVisible();

    // Delete User
    page.on('dialog', dialog => dialog.accept());

    // Mock Delete and subsequent List
    await page.route(`**/api/v1/users/${USER_ID}`, async route => {
         await route.fulfill({ json: {} });
    });
     await page.route('**/api/v1/users', async route => {
        if (route.request().method() === 'GET') {
             await route.fulfill({ json: { users: [] } }); // Gone
        }
    });

    await row.getByRole('button').nth(1).click(); // Second button is Delete (Trash)

    // Verify deletion
    await expect(page.getByRole('cell', { name: USER_ID })).not.toBeVisible();
  });
});
