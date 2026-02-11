/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Onboarding Flow', () => {
  test('shows welcome wizard when no services exist', async ({ page }) => {
    // Mock empty services list
    await page.route('**/api/v1/services', async route => {
      await route.fulfill({ json: { services: [] } });
    });

    await page.goto('/');

    // Check for Welcome Wizard elements
    await expect(page.getByText('Welcome to MCP Any')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Get Started' })).toBeVisible();
  });

  test('shows dashboard when services exist', async ({ page }) => {
    // Mock services list with one service
    await page.route('**/api/v1/services', async route => {
      await route.fulfill({
        json: {
          services: [
            {
              name: 'test-service',
              id: 'test-service',
              httpService: { address: 'http://localhost:8080' }
            }
          ]
        }
      });
    });

    await page.goto('/');

    // Check for Dashboard elements
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
    // Wizard should NOT be visible
    await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible();
  });

  test('wizard progression', async ({ page }) => {
    // Start with empty services
    await page.route('**/api/v1/services', async (route, request) => {
        if (request.method() === 'GET') {
             await route.fulfill({ json: { services: [] } });
        } else if (request.method() === 'POST') {
             // Mock registration success
             await route.fulfill({ json: { name: 'new-service' } });
        } else {
            await route.continue();
        }
    });

    // Mock validation
    await page.route('**/api/v1/services/validate', async route => {
        await route.fulfill({ json: { valid: true, message: 'Valid' } });
    });

    await page.goto('/');

    // Step 1: Click Get Started
    await page.getByRole('button', { name: 'Get Started' }).click();

    // Step 2: Register Service View
    await expect(page.getByText('Step 1: Register a Service')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Register New Service' })).toBeVisible();

    // Open Register Dialog
    await page.getByRole('button', { name: 'Register New Service' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    // Select "Custom Service" template (if templates view is active)
    await page.getByText('Custom Service').click();

    // Fill Form (Basic HTTP)
    await page.locator('input[name="name"]').fill('my-new-service');
    // Select type if needed (default is HTTP)
    await page.locator('input[name="address"]').fill('https://example.com');

    // Validate
    await page.getByRole('button', { name: 'Test Connection' }).click();
    // Expect specific success message, scoping to dialog if possible or just first/exact
    await expect(page.getByText('Validation Successful').first()).toBeVisible();

    // Save
    await page.getByRole('button', { name: 'Register Service' }).click();

    // Should advance to Step 3: Connect Client
    await expect(page.getByText('Step 2: Connect AI Client')).toBeVisible();

    // Step 3: Complete
    await page.getByRole('button', { name: 'Go to Dashboard' }).click();

    // Should call onComplete -> Dashboard
    // Since we didn't update the GET mock to return services, the component state 'hasServices'
    // should have been set to true by the callback.
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
  });
});
