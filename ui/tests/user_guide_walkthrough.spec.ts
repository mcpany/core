/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('User Guide Walkthrough', () => {
  test('Dashboard loads key metrics', async ({ page }) => {
    // Mock the stats endpoint
    await page.route('**/api/v1/dashboard/metrics', async route => {
        await route.fulfill({
            json: [
                { label: "Total Requests", value: "1234", icon: "Activity" },
                { label: "Active Services", value: "5", icon: "Server" },
                { label: "Connected Tools", value: "12", icon: "Zap" }
            ]
        });
    });

    await page.goto('/');
    // Check for "Total Requests" card
    await expect(page.locator('text=Total Requests')).toBeVisible({ timeout: 10000 });
    // Check for "Active Services" card
    await expect(page.locator('text=Active Services')).toBeVisible();
    await expect(page.locator('text=Connected Tools')).toBeVisible();
  });

  test('Services: Add Service Redirects to Marketplace', async ({ page }) => {
    await page.goto('/upstream-services');
    await expect(page.getByRole('heading', { name: 'Upstream Services' })).toBeVisible();

    // Explicitly target the button with text "Add Service"
    const addButton = page.getByRole('button', { name: 'Add Service' });
    await expect(addButton).toBeVisible();

    // Check for dialog opens
    await addButton.click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('New Service')).toBeVisible();

    // Close it
    await page.keyboard.press('Escape');
    await expect(page.getByRole('dialog')).toBeHidden();
  });

  test('Resources: List and Preview', async ({ page }) => {
    // Mock resources to ensure 'config.json' is present
    await page.route('**/api/v1/resources', async route => {
        const json = {
            resources: [{
                uri: 'config.json',
                name: 'config.json',
                mimeType: 'application/json'
            }]
        };
        await route.fulfill({ json });
    });

    await page.goto('/resources');
    await expect(page.getByRole('heading', { name: 'Resources' })).toBeVisible();

    // Check for known mock resources from walkthrough or just non-empty page
    // "config.json" was seen in verification
    await expect(page.locator('body')).toContainText('config.json');
  });

  test('Global Search Modal', async ({ page }) => {
    await page.goto('/');

    // Press Ctrl+K
    await page.keyboard.press('Control+k');

    // Check for Search Input placeholder or identifying element
    await expect(page.getByPlaceholder('Type a command or search...')).toBeVisible({ timeout: 20000 });

    // Close modal (Esc key or click outside/close button)
    await page.keyboard.press('Escape');

    await expect(page.getByPlaceholder('Type a command or search...')).not.toBeVisible();
  });

  test('Logs Stream', async ({ page }) => {
    await page.goto('/logs');
    await expect(page.getByRole('heading', { name: 'Live Logs' })).toBeVisible();
    // Check for log container - using more specific selector to avoid strict mode violation
    // expected container has bg-black/90
    await expect(page.locator('div.font-mono.bg-black\\/90')).toBeVisible();
  });

  test('Secrets Vault', async ({ page }) => {
    await page.goto('/secrets');
    await expect(page.getByRole('heading', { name: 'API Keys & Secrets' })).toBeVisible();
    // Check for Add Secret button
    await expect(page.getByRole('button', { name: 'Add Secret' })).toBeVisible();
  });

  test('Alerts Page', async ({ page }) => {
    await page.goto('/alerts');
    await expect(page.getByRole('heading', { name: 'Alerts & Incidents' })).toBeVisible();
    // Check for "Alerts & Incidents" text or stats
    await expect(page.getByText('Monitor system health')).toBeVisible();
  });

  test('Stack Composer', async ({ page }) => {
    await page.goto('/stacks');
    await expect(page.getByRole('heading', { name: 'Stacks' })).toBeVisible();
    // Check for "Create Stack" button which is now implemented
    await expect(page.getByRole('button', { name: 'Create Stack' })).toBeVisible();
  });

  test('Webhooks Management', async ({ page }) => {
    await page.goto('/webhooks');
    await expect(page.getByRole('heading', { name: 'Webhooks' })).toBeVisible();
    // Button is "New Webhook", not "Add Webhook"
    await expect(page.getByRole('button', { name: 'New Webhook' })).toBeVisible();
  });

  test('Connection Diagnostic Tool', async ({ page }) => {
    // Mock services response to ensure at least one service exists
    await page.route('**/api/v1/services', async route => {
      const json = [{
        name: 'mock-service',
        id: 'mock-service-id',
        http_service: { address: 'http://localhost:8080' },
        disable: false
      }];
      await route.fulfill({ json });
    });

    // Mock getService response as well since clicking navigates to detail which fetches /api/v1/services/mock-service-id
    await page.route('**/api/v1/services/mock-service-id', async route => {
        const json = {
            service: {
                name: 'mock-service',
                id: 'mock-service-id',
                http_service: { address: 'http://localhost:8080' },
                disable: false,
                version: '1.0.0'
            }
        };
        await route.fulfill({ json });
    });

    // Navigate to services first
    await page.goto('/upstream-services');
    // Open Edit Sheet
    const row = page.locator('tr').filter({ hasText: 'mock-service' });
    await expect(row).toBeVisible();
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Check for Edit Sheet load
    await expect(page.getByRole('heading', { name: 'Edit Service' })).toBeVisible();
    await expect(page.locator('input[id="name"]')).toHaveValue('mock-service');
  });
});
