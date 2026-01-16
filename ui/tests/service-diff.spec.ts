import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';

test.describe('Service Config Diff', () => {
  test('should show diff when editing a service', async ({ page }) => {
    // Mock the service response
    const service = {
      id: 's1',
      name: 'Test Service',
      version: '1.0.0',
      disable: false,
      httpService: {
        address: 'http://localhost:8080',
        tools: [],
        calls: {},
        resources: [],
        prompts: []
      }
    };

    // Abort gRPC calls to force fallback to REST
    await page.route('**/mcpany.api.v1.RegistrationService/**', async route => {
        await route.abort();
    });

    // Mock REST API endpoints - Catch ALL service requests (GET, PUT, POST)
    await page.route('**/api/v1/services**', async route => {
        const method = route.request().method();
        const url = route.request().url();
        console.log(`Intercepted ${method} ${url}`);

        if (method === 'GET') {
            // Check if specific service
             if (url.includes('/s1') || url.includes('/Test%20Service')) {
                 await route.fulfill({ json: { service } });
             } else {
                 // List
                 await route.fulfill({ json: { services: [service] } });
             }
        } else if (method === 'POST' || method === 'PUT') {
             // Mock update/create response
             // We can return the updated service
             await route.fulfill({ json: { service: { ...service, name: 'Test Service Updated' } } });
        } else {
             await route.continue();
        }
    });

    await page.route('**/api/v1/credentials', async route => {
        await route.fulfill({ json: [] });
    });

    // Mock doctor to avoid noise
    await page.route('**/doctor', async route => {
        await route.fulfill({ json: { status: 'healthy', version: '1.0.0' } });
    });

     // Mock connection diagnostics
    await page.route('**/api/v1/diagnostics/**', async route => {
        await route.fulfill({ json: { status: 'ok' } });
    });

    // Go to service detail page
    await page.goto('/service/s1');

    // Wait for page to load
    await expect(page.getByText('Test Service').first()).toBeVisible();

    // Click Edit Config
    await page.getByRole('button', { name: 'Edit Config' }).click();

    // Wait for dialog
    await expect(page.getByRole('dialog')).toBeVisible();

    // Change name
    await page.getByLabel('Service Name').fill('Test Service Updated');

    // Click Review Changes
    await page.getByRole('button', { name: 'Review Changes' }).click();

    // Verify Diff View
    await expect(page.getByText('Review Changes', { exact: true }).first()).toBeVisible();

    // Verify Back and Confirm buttons
    await expect(page.getByRole('button', { name: 'Back to Edit' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Confirm & Save' })).toBeVisible();

    // Screenshot
    const date = new Date().toISOString().split('T')[0];
    const dir = path.resolve(process.cwd(), `../.audit/ui/${date}`);
    if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
    }
    await page.screenshot({ path: path.join(dir, 'service_config_diff_viewer.png') });

    // Click Confirm
    await page.getByRole('button', { name: 'Confirm & Save' }).click();

    // Verify success toast or dialog closed
    await expect(page.getByRole('dialog')).not.toBeVisible();
    await expect(page.getByText('Service Updated').first()).toBeVisible();
  });
});
