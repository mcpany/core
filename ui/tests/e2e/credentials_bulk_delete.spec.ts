/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Credentials Bulk Delete', () => {
  const creds = [
    { name: 'e2e-cred-1', type: 'api_key', apiKey: { paramName: 'X-API-Key', in: 0, value: 'secret1' } },
    { name: 'e2e-cred-2', type: 'bearer_token', bearerToken: { token: 'secret2' } },
    { name: 'e2e-cred-3', type: 'basic_auth', basicAuth: { username: 'user', password: 'password' } },
  ];

  const headers = { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token', 'Content-Type': 'application/json' };
  const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';

  test.beforeEach(async ({ playwright }) => {
    const context = await playwright.request.newContext({ baseURL: BASE_URL });

    // Wait for API to be ready
    await expect(async () => {
        const res = await context.get('/health', { headers });
        expect(res.ok()).toBeTruthy();
    }).toPass({ timeout: 10000 });

    // Clean up first just in case
    const listRes = await context.get('/api/v1/credentials', { headers });
    if (listRes.ok()) {
      const data = await listRes.json();
      const credentials = Array.isArray(data) ? data : (data.credentials || []);
      for (const cred of credentials) {
        if (creds.some(c => c.name === cred.name)) {
          await context.delete(`/api/v1/credentials/${cred.id}`, { headers });
        }
      }
    }

    // Seed the database with test credentials
    for (const cred of creds) {
      let data: any = { name: cred.name };
      if (cred.type === 'api_key') {
        data.authentication = { api_key: { param_name: cred.apiKey.paramName, in: cred.apiKey.in, value: { plain_text: cred.apiKey.value } } };
      } else if (cred.type === 'bearer_token') {
        data.authentication = { bearer_token: { token: { plain_text: cred.bearerToken.token } } };
      } else if (cred.type === 'basic_auth') {
        data.authentication = { basic_auth: { username: cred.basicAuth.username, password: { plain_text: cred.basicAuth.password } } };
      }

      const r = await context.post('/api/v1/credentials', { headers, data });
      expect(r.ok()).toBeTruthy();
    }
    await context.dispose();
  });

  test.afterEach(async ({ playwright }) => {
    const context = await playwright.request.newContext({ baseURL: BASE_URL });
    // Clean up just in case
    const response = await context.get('/api/v1/credentials', { headers });
    if (response.ok()) {
      const data = await response.json();
      const credentials = Array.isArray(data) ? data : (data.credentials || []);
      for (const cred of credentials) {
        if (creds.some(c => c.name === cred.name)) {
          await context.delete(`/api/v1/credentials/${cred.id}`, { headers });
        }
      }
    }
    await context.dispose();
  });

  test('should display credentials, allow bulk selection, and bulk delete them', async ({ page }) => {
    await page.goto('/credentials');

    // Wait for the credentials to load
    await expect(page.getByText('e2e-cred-1')).toBeVisible();
    await expect(page.getByText('e2e-cred-2')).toBeVisible();
    await expect(page.getByText('e2e-cred-3')).toBeVisible();

    // Verify "Delete Selected" button is hidden initially
    const deleteSelectedBtn = page.getByRole('button', { name: /Delete Selected/i });
    await expect(deleteSelectedBtn).toBeHidden();

    // Click the "Select All" checkbox in the table header
    const selectAllCheckbox = page.getByRole('checkbox', { name: 'Select all' });
    await selectAllCheckbox.click();

    // Wait for the "Delete Selected" button to appear
    await expect(deleteSelectedBtn).toBeVisible();

    // Handle confirmation dialog
    page.on('dialog', async dialog => {
        expect(dialog.message()).toContain('delete');
        await dialog.accept();
    });

    // Click "Delete Selected"
    await deleteSelectedBtn.click();

    // Wait for the deletion to complete and table to update
    await expect(page.getByText('e2e-cred-1')).toBeHidden();
    await expect(page.getByText('e2e-cred-2')).toBeHidden();
    await expect(page.getByText('e2e-cred-3')).toBeHidden();
  });
});
