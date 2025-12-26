/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


/**
 * End-to-end tests for the MCP Manager UI.
 * Verifies critical user flows including dashboard metrics, service management,
 * and tools/resources/prompts exploration.
 */

import { test, expect } from '@playwright/test';
import * as path from 'path';

test.describe('MCP Manager E2E', () => {

  /**
   * Verifies that the dashboard loads correctly and displays key metrics and the system health widget.
   */
  test('Dashboard loads with metrics and health widget', async ({ page }) => {
    await page.goto('/');

    // Check for dashboard title (this is h2 in page.tsx)
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

    // Check for metrics - wait for loading
    // The loading state has "Loading dashboard metrics..."
    await expect(page.getByText('Loading dashboard metrics...')).not.toBeVisible({ timeout: 10000 });

    await expect(page.getByText('Total Requests')).toBeVisible();
    await expect(page.getByText('Active Services')).toBeVisible();

    // Using locator with specific class to distinguish from other occurrences
    // The metric card title has "text-sm font-medium"
    await expect(page.locator('.text-sm.font-medium', { hasText: 'Resources' })).toBeVisible();
    await expect(page.locator('.text-sm.font-medium', { hasText: 'Prompts' })).toBeVisible();

    // Check for health widget
    await expect(page.getByText('System Health').first()).toBeVisible();
    await expect(page.getByText('Core API Gateway')).toBeVisible(); // Mocked service

    // Take screenshot
    await page.screenshot({ path: '.audit/ui/2025-05-23/dashboard.png', fullPage: true });
  });

  /**
   * Verifies that the services page lists registered services and allows toggling their status.
   */
  test('Services page lists services and allows toggling', async ({ page }) => {
    await page.goto('/services');

    await expect(page.getByRole('heading', { name: 'Services' })).toBeVisible();
    // CardTitle is a div
    await expect(page.getByText('Upstream Services', { exact: true })).toBeVisible();

    // Check for mock service
    await expect(page.getByRole('cell', { name: 'weather-service' })).toBeVisible();

    // Toggle service (optimistic update check)
    const row = page.getByRole('row', { name: 'weather-service' });
    const toggle = row.getByRole('switch');

    await expect(toggle).toBeChecked(); // Initially active
    await toggle.click();
    await expect(toggle).not.toBeChecked();

    await page.screenshot({ path: '.audit/ui/2025-05-23/services.png', fullPage: true });
  });

  /**
   * Verifies that the tools page lists available tools.
   */
  test('Tools page lists tools', async ({ page }) => {
    await page.goto('/tools');

    await expect(page.getByRole('heading', { name: 'Tools' })).toBeVisible();
    await expect(page.getByText('Available Tools', { exact: true })).toBeVisible();

    await expect(page.getByRole('cell', { name: 'get_weather' })).toBeVisible();

    await page.screenshot({ path: '.audit/ui/2025-05-23/tools.png', fullPage: true });
  });

  /**
   * Verifies that the resources page lists available resources.
   */
  test('Resources page lists resources', async ({ page }) => {
    await page.goto('/resources');

    await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible();
    await expect(page.getByText('Managed Resources', { exact: true })).toBeVisible();

    // notes.txt appears in Name and URI. We want the one in the first column (Name)
    // or just check that *some* cell has it exactly.
    await expect(page.getByRole('cell', { name: 'notes.txt', exact: true })).toBeVisible();

    await page.screenshot({ path: '.audit/ui/2025-05-23/resources.png', fullPage: true });
  });

  /**
   * Verifies that the prompts page lists available prompt templates.
   */
  test('Prompts page lists prompts', async ({ page }) => {
    await page.goto('/prompts');

    await expect(page.getByRole('heading', { name: 'Prompts' })).toBeVisible();
    await expect(page.getByText('Prompt Templates', { exact: true })).toBeVisible();

    await expect(page.getByRole('cell', { name: 'summarize_file' })).toBeVisible();

    await page.screenshot({ path: '.audit/ui/2025-05-23/prompts.png', fullPage: true });
  });

  /**
   * Verifies that the middleware page loads and displays the pipeline visualization.
   */
  test('Middleware page interaction', async ({ page }) => {
      await page.goto('/middleware');

      await expect(page.getByRole('heading', { name: 'Middleware Pipeline' })).toBeVisible();

      await expect(page.getByText('Authentication').first()).toBeVisible();

      // Verify visual pipeline renders
      await expect(page.getByText('Pipeline Visualization', { exact: true })).toBeVisible();

      await page.screenshot({ path: '.audit/ui/2025-05-23/middleware.png', fullPage: true });
  });

  /**
   * Verifies that the webhooks page loads and allows opening the "New Webhook" modal.
   */
  test('Webhooks page interaction', async ({ page }) => {
      await page.goto('/webhooks');

      await expect(page.getByRole('heading', { name: 'Webhooks' })).toBeVisible();

      // Click New Webhook
      await page.getByRole('button', { name: 'New Webhook' }).click();
      await expect(page.getByRole('heading', { name: 'Add Webhook' })).toBeVisible();

      await page.screenshot({ path: '.audit/ui/2025-05-23/webhooks.png', fullPage: true });
  });
});
