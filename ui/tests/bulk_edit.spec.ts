/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

test.describe('Bulk Edit Services', () => {
  test.beforeEach(async ({ request, page }) => {
      // Clean up first to ensure clean state
      await cleanupServices(request);
      await cleanupUser(request, "admin");

      // Seed standard services
      await seedServices(request);

      // Seed a CLI service for Env Var testing
      const cliService = {
        id: "svc_cli_01",
        name: "CLI Tool",
        version: "1.0.0",
        command_line_service: {
            command: "echo",
            env: { "EXISTING_VAR": { plain_text: "value" } }
        }
      };
      await request.post('/api/v1/services', {
          data: cliService,
          headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
      });

      // Seed user and login
      await seedUser(request, "admin");

      await page.goto('/login');
      // Wait for login form
      await expect(page.locator('input[name="username"]')).toBeVisible();
      await page.fill('input[name="username"]', 'admin');
      await page.fill('input[name="password"]', 'password');
      await page.click('button[type="submit"]');

      // Wait for redirect to home
      await expect(page).toHaveURL('/', { timeout: 15000 });

      // Navigate to services page
      await page.goto('/upstream-services');
      // Wait for table to load
      await expect(page.getByRole('heading', { name: 'Upstream Services' })).toBeVisible();
      await expect(page.locator('tr', { hasText: 'Payment Gateway' })).toBeVisible();
  });

  test.afterEach(async ({ request }) => {
      await cleanupServices(request);
      await request.delete('/api/v1/services/CLI Tool', {
          headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
      });
      await cleanupUser(request, "admin");
  });

  test('should bulk edit tags, timeout, and env vars', async ({ page, request }) => {
      // 1. Select Services
      await page.getByLabel('Select Payment Gateway').check();
      await page.getByLabel('Select CLI Tool').check();

      // 2. Click Bulk Edit
      await expect(page.getByRole('button', { name: 'Bulk Edit' })).toBeVisible();
      await page.getByRole('button', { name: 'Bulk Edit' }).click();

      // 3. Fill Form
      await expect(page.getByRole('dialog')).toBeVisible();
      await expect(page.getByRole('heading', { name: 'Bulk Edit Services' })).toBeVisible();

      // Add Tags
      await page.getByLabel('Add Tags').fill('bulk-tag');

      // Set Timeout
      await page.getByLabel('Set Timeout').fill('30s');

      // Add Env Var
      await page.getByPlaceholder('Key').fill('NEW_VAR');
      await page.getByPlaceholder('Value').fill('new_value');
      await page.getByRole('button', { name: 'Add Env' }).click();

      // 4. Apply
      await page.getByRole('button', { name: 'Apply Changes' }).click();

      // 5. Verify Toast
      await expect(page.getByText('Services Updated').first()).toBeVisible();

      // 6. Verify Tag in UI
      await page.reload();
      await expect(page.locator('tr', { hasText: 'Payment Gateway' })).toContainText('bulk-tag');
      await expect(page.locator('tr', { hasText: 'CLI Tool' })).toContainText('bulk-tag');

      // 7. Verify via API for Timeout and Env Vars
      // Verify CLI Tool
      const resCli = await request.get('/api/v1/services/CLI Tool', {
          headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
      });
      const dataCli = await resCli.json();

      // Verify Timeout
      // Ensure resilience exists
      // Note: Response structure is the service object itself, not wrapped in { service: ... } for direct REST call?
      // Based on debug log, it is direct object.
      // But apiClient.getService expects { service: ... } wrapper?
      // Or maybe apiClient logic handles both.
      // Let's assume direct object based on logs.
      const service = dataCli.service || dataCli;

      expect(service.resilience).toBeDefined();
      expect(service.resilience.timeout).toBe('30s');

      // Verify Env Vars
      // command_line_service.env should contain NEW_VAR and EXISTING_VAR
      // Accessing dynamic keys safely
      const env = service.command_line_service.env;
      expect(env).toBeDefined();
      expect(env['NEW_VAR'].plain_text).toBe('new_value');
      // Existing var should still be there (even if redacted/empty value returned by API)
      expect(env['EXISTING_VAR']).toBeDefined();

      // Verify Payment Gateway
      const resPg = await request.get('/api/v1/services/Payment Gateway', {
          headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
      });
      const dataPg = await resPg.json();
      const servicePg = dataPg.service || dataPg;

      // Verify Timeout
      expect(servicePg.resilience.timeout).toBe('30s');

      // Verify Env - HTTP service should NOT have env vars added
      expect(servicePg.command_line_service).toBeUndefined();
  });
});
