/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

const DOCS_SCREENSHOTS_DIR = path.resolve(__dirname, '../docs/screenshots');

if (!fs.existsSync(DOCS_SCREENSHOTS_DIR)) {
  fs.mkdirSync(DOCS_SCREENSHOTS_DIR, { recursive: true });
}

test.describe('Generate Detailed Docs Screenshots', () => {

  test.beforeEach(async ({ page }) => {
     // No mocks - using real backend
  });

  test('Dashboard Screenshots', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);
    await expect(page.locator('body')).toBeVisible();
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'dashboard_overview.png'), fullPage: true });
    // Legacy alias
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'dashboard.png'), fullPage: true });
  });

  test('Services Screenshots', async ({ page }) => {
    await page.goto('/upstream-services');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    // Wait for loading to finish if applicable
    await expect(page.locator('text=Loading...')).not.toBeVisible();

    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services_list.png'), fullPage: true });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services.png'), fullPage: true });

    // Click Add Service (Link)
    await page.getByRole('link', { name: 'Add Service' }).click();
    await page.waitForTimeout(1000);
    await expect(page).toHaveURL(/.*marketplace.*/);

    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'services_add_dialog.png') });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'service_add_dialog.png') }); // Alias

    // Configure Service
    await page.goto('/upstream-services/postgres-primary');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'service_config.png'), fullPage: true });
  });

  test('Playground Screenshots', async ({ page }) => {
    await page.goto('/playground');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground_blank.png'), fullPage: true });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground.png'), fullPage: true });

    // No mocks - using real backend

    // Reload to get tools
    await page.reload();
    await page.waitForTimeout(1000);

    // Select Tool
    const tool = page.getByText('filesystem.list_dir');
    if (await tool.isVisible()) {
        await tool.click();
        await page.waitForTimeout(500);
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground_tool_selected.png'), fullPage: true });

        // Fill Form
        await page.getByLabel('path').fill('/var/log');
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'playground_form_filled.png'), fullPage: true });
    }

    // Tools alias
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'tools.png'), fullPage: true });
  });

  test('Stack Composer Screenshots', async ({ page }) => {
    await page.goto('/stacks');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stack_composer_overview.png'), fullPage: true });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stacks.png'), fullPage: true });

    if (await page.getByText('Service Palette').isVisible()) {
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stack_composer_palette.png'), fullPage: true });
    }
  });

  test('Traces Screenshots', async ({ page }) => {
    // No mocks - using real backend

    await page.goto('/traces');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'traces_list.png'), fullPage: true });
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'traces.png'), fullPage: true });

    // Click trace
    await page.getByText('filesystem.read').first().click({ force: true });
    await page.waitForTimeout(500);
    await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'trace_detail.png'), fullPage: true });
  });

  test('Middleware Screenshots', async ({ page }) => {
      await page.goto('/middleware');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'middleware_pipeline.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'middleware.png'), fullPage: true });
  });

  test('Webhooks Screenshots', async ({ page }) => {
      await page.goto('/webhooks');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'webhooks_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'webhooks.png'), fullPage: true });
      // Legacy alias
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_webhooks.png'), fullPage: true });

      await page.getByText('New Webhook').click();
      await page.waitForTimeout(500);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'webhook_create_modal.png') });
  });

  test('Network Graph Screenshots', async ({ page }) => {
      await page.goto('/network');
      await page.waitForTimeout(2000); // Graph rendering
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'network_graph.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'network.png'), fullPage: true });
  });

  test('Logs Screenshots', async ({ page }) => {
      await page.goto('/logs');
      await page.waitForTimeout(1000);
       await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'logs_stream.png'), fullPage: true });
       await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'logs.png'), fullPage: true });
  });

   test('Marketplace Screenshots', async ({ page }) => {
       await page.goto('/marketplace');
       await page.waitForTimeout(1000);
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'marketplace_grid.png'), fullPage: true });
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'marketplace.png'), fullPage: true });
        // Legacy alias
        await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'marketplace_external_detailed.png'), fullPage: true });
   });

  test('Secrets Screenshots', async ({ page }) => {
      // No mocks - using real backend
      await page.goto('/secrets');
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1000);
      await expect(page.getByText('Loading secrets...')).not.toBeVisible();
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secrets_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secrets.png'), fullPage: true });
      // Legacy alias (Secrets List)
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_secrets.png'), fullPage: true });

      await page.getByText('Add Secret').click();
      await page.waitForTimeout(500);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'secret_create_modal.png') });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'credential_form.png') });
  });

  test('Auth Screenshots', async ({ page }) => {
      await page.goto('/login');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_login.png'), fullPage: true });
      // Legacy aliases (placeholder to ensure update)
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_guide_step1_apikey.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_guide_step2_bearer.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_guide_step3_basic.png'), fullPage: true });

      // Mock Users
      await page.goto('/users');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'auth_users_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'users.png'), fullPage: true });
  });

  test('Prompts Screenshots', async ({ page }) => {
      await page.goto('/prompts');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'prompts_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'prompts.png'), fullPage: true });
  });

  test('Search Screenshots', async ({ page }) => {
       await page.goto('/');
       await page.waitForTimeout(1000);
       await page.keyboard.press('Control+k');
       await page.waitForTimeout(500);
       await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'global_search.png') });
  });

  test('Resources Screenshots', async ({ page }) => {
      await page.goto('/resources');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resources_list.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resources.png'), fullPage: true });
      // Legacy aliases
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resources_grid.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'resources_split_view.png'), fullPage: true });
  });

  test('Alerts Screenshots', async ({ page }) => {
      await page.goto('/alerts');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'alerts_list.png'), fullPage: true });
  });

  test('Mobile Screenshots', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 812 });
      await page.goto('/');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'mobile_dashboard.png'), fullPage: true });
  });

  test('Skills Screenshots', async ({ page }) => {
      await page.goto('/skills');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'skills_list.png'), fullPage: true });

      // Create/Edit View
      await page.goto('/skills/create');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'skills_create.png'), fullPage: true });
  });

  test('Settings Screenshots', async ({ page }) => {
      await page.goto('/settings');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_profiles.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'profiles.png'), fullPage: true }); // Legacy alias

      // Click General Tab
      await page.getByRole('tab', { name: 'General' }).click();
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_general.png'), fullPage: true });

      // Legacy alias for Auth Settings (which is another tab)
      await page.getByRole('tab', { name: 'Authentication' }).click();
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'settings_auth.png'), fullPage: true });
  });

  test('Credentials Screenshots', async ({ page }) => {
      await page.goto('/credentials');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'credentials.png'), fullPage: true });
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'credentials_list.png'), fullPage: true });
  });

  test('Stats Screenshots', async ({ page }) => {
      await page.goto('/stats');
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(DOCS_SCREENSHOTS_DIR, 'stats.png'), fullPage: true });
  });

});
