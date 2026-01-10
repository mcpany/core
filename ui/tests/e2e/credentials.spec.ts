/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Credentials Management', () => {



  test('should list, create, update and delete credentials', async ({ page }) => {
    // 1. Initial List (Empty)
    await page.route('**/api/v1/credentials', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({ json: [] });
      } else {
        await route.continue();
      }
    });

    await page.goto('/credentials');
    await expect(page.getByText('No credentials found')).toBeVisible();

    // 2. Create Credential
    const newCred = {
      id: 'cred-1',
      name: 'Test API Key',
      authentication: { apiKey: { paramName: 'Authorization', in: 0, value: { plainText: 'secret-key' } } }
    };

    let created = false;
    await page.route('**/api/v1/credentials', async route => {
      const method = route.request().method();
      if (method === 'POST') {
        created = true;
        await route.fulfill({ json: newCred });
      } else if (method === 'GET') {
        await route.fulfill({ json: created ? [newCred] : [] });
      } else {
        await route.continue();
      }
    });

    await page.getByRole('button', { name: 'New Credential' }).click();
    await expect(page.getByText('Create Credential')).toBeVisible();

    await page.getByPlaceholder('My Credential').fill('Test API Key');
    // Default format is API Key, so just fill details
    await page.getByPlaceholder('X-API-Key').fill('Authorization');
    await page.getByPlaceholder('...secret key...').fill('secret-key');

    await page.getByRole('button', { name: 'Save' }).click();

    // Verify it appears in list
    await expect(page.getByText('Test API Key')).toBeVisible();
    await expect(page.locator('tbody').getByText('API Key', { exact: true })).toBeVisible();

    // 3. Update Credential
    await page.route(`**/api/v1/credentials/${newCred.id}`, async route => {
        if (route.request().method() === 'PUT') {
             const data = route.request().postDataJSON();
             newCred.name = data.name;
             await route.fulfill({ json: newCred });
        } else {
             await route.continue();
        }
    });

    // Refresh mock for list to return updated name
    await page.route('**/api/v1/credentials', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({ json: [newCred] });
        } else {
            await route.continue();
        }
    });

    await page.getByRole('button', { name: 'Edit' }).click();
    await page.getByPlaceholder('My Credential').fill('Updated API Key');
    await page.getByRole('button', { name: 'Save' }).click();

    await expect(page.getByText('Updated API Key')).toBeVisible();

    // 4. Delete Credential
    await page.route(`**/api/v1/credentials/${newCred.id}`, async route => {
        if (route.request().method() === 'DELETE') {
             await route.fulfill({ status: 200 });
        }
    });

    // Refresh mock for list to return empty
    await page.route('**/api/v1/credentials', async route => {
        if (route.request().method() === 'GET') {
            await route.fulfill({ json: [] });
        }
    });

    // Accept delete confirmation
    page.on('dialog', dialog => dialog.accept());

    await page.getByRole('button', { name: 'Delete' }).click();
    // In our UI, we might use a custom dialog instead of window.confirm
    // If it's a Radix alert dialog:
    if (await page.getByText('Are you sure?').isVisible()) {
         await page.getByRole('button', { name: 'Delete' }).last().click();
    }

    await expect(page.getByText('No credentials found')).toBeVisible();
  });
});
