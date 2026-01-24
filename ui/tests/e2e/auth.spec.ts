/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

// Use a unique ID to avoid collisions
const TEST_TIMESTAMP = Date.now();
const USER_ID = `e2e-user-${TEST_TIMESTAMP}`;

test.describe('Authentication and User Management', () => {

  test('should login successfully with real backend', async ({ page, request }) => {
      // 1. Create User via API (using Admin/API Key from config)
      // Inject API Key header explicitly since we removed auto-injection from middleware
      const apiKey = process.env.MCPANY_API_KEY || 'test-token';

      const createUserRes = await request.post('/api/v1/users', {
          headers: {
              'X-API-Key': apiKey
          },
          data: {
              user: {
                  id: USER_ID,
                  roles: ['viewer'],
                  authentication: {
                      basic_auth: {
                          username: USER_ID,
                          password_hash: 'password123' // Server handles hashing
                      }
                  }
              }
          }
      });

      // If user creation fails, test should fail
      if (!createUserRes.ok()) {
          console.error("Failed to create user:", await createUserRes.text());
      }
      expect(createUserRes.ok()).toBeTruthy();

      // 2. Login via UI
      await page.goto('/login');
      // Ensure we are logically on login page
      await expect(page.getByLabel('Username')).toBeVisible();

      await page.getByLabel('Username').fill(USER_ID);
      await page.getByLabel('Password').fill('password123');
      await page.getByRole('button', { name: 'Sign in' }).click();

      // 3. Verify Redirect to Dashboard
      await expect(page).toHaveURL('/');

      // 4. Verify Token in LocalStorage
      const token = await page.evaluate(() => localStorage.getItem('mcp_auth_token'));
      expect(token).toBeTruthy();

      // 5. Verify User Management Access (Protected)
      await page.goto('/users');
      await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();

      // Verify our user is in the list
      await expect(page.getByRole('cell', { name: USER_ID })).toBeVisible();

      // Cleanup: Delete user
      const deleteRes = await request.delete(`/api/v1/users/${USER_ID}`, {
          headers: {
              'X-API-Key': apiKey
          }
      });
      expect(deleteRes.ok()).toBeTruthy();
  });
});
