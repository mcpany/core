/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Alerts Page', () => {
  // Use beforeAll for data setup, but note that server state might persist or reset depending on CI
  // We use unique prefixes or reset logic if needed. Here we assume a fresh server or we can handle existing data.
  // Actually, beforeAll hook in Playwright runs once per worker.
  test.beforeEach(async ({ request }) => {
    // Seed required alerts if not present. Idempotency is key.
    // Since we don't have unique constraints on title, we might create duplicates if we run multiple times.
    // But for this test, we can just create them and filter by unique text we look for.
    await request.post('/api/v1/alerts', {
      data: {
        title: "High CPU Usage",
        message: "CPU usage > 90%",
        severity: "critical",
        status: "active",
        service: "test-service",
        source: "test"
      }
    });
    await request.post('/api/v1/alerts', {
      data: {
        title: "API Latency Spike",
        message: "Latency high",
        severity: "warning",
        status: "active",
        service: "api-gateway",
        source: "test"
      }
    });
    await request.post('/api/v1/alerts', {
        data: {
          title: "Disk Space Low",
          message: "Disk space low",
          severity: "warning",
          status: "active",
          service: "db-service",
          source: "test"
        }
      });
  });

  test('should load alerts page and display key elements', async ({ page }) => {
    // Navigate to alerts page
    await page.goto('/alerts');

    // Check header
    await expect(page.getByRole('heading', { name: 'Alerts & Incidents' })).toBeVisible();

    // Check stats cards
    await expect(page.getByText('Active Critical')).toBeVisible();
    await expect(page.getByText('MTTR (Today)')).toBeVisible();

    // Check table content (seeded data)
    await expect(page.getByText('High CPU Usage')).toBeVisible();
    await expect(page.getByText('API Latency Spike')).toBeVisible();
  });

  test('should filter alerts', async ({ page }) => {
    await page.goto('/alerts');

    // Type in search box - use getByPlaceholder if available, else locator
    const searchBox = page.locator('input[placeholder="Search alerts by title, message, service..."]');
    await searchBox.fill('CPU');

    // Should see CPU alert
    await expect(page.getByText('High CPU Usage')).toBeVisible();

    // Should NOT see Latency alert
    await expect(page.getByText('API Latency Spike')).toBeHidden();
  });

  test('should open create rule dialog', async ({ page }) => {
    await page.goto('/alerts');

    // Click "New Alert Rule" button
    await page.getByRole('button', { name: 'New Alert Rule' }).click();

    // Check dialog opens
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Create Alert Rule' })).toBeVisible();

    // Close it
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('dialog')).toBeHidden();
  });
  test('should acknowledge alert via dropdown', async ({ page }) => {
    await page.goto('/alerts');

    // Find an active alert row (mock data usually has some)
    // We target the row with "High CPU Usage" which is active in mock
    const row = page.getByRole('row').filter({ hasText: 'High CPU Usage' });

    // Click the "More Actions" dropdown button in that row
    await row.getByRole('button', { name: 'Open menu' }).click();

    // Click "Acknowledge"
    await page.getByRole('menuitem', { name: 'Acknowledge' }).click();

    // Verify status changes to "acknowledged"
    await expect(row.getByText('acknowledged')).toBeVisible();
  });

  test('should resolve alert via dropdown', async ({ page }) => {
    await page.goto('/alerts');

    // Find an acknowledged or active alert
    const row = page.getByRole('row').filter({ hasText: 'Disk Space Low' });

    // Click "More Actions"
    await row.getByRole('button', { name: 'Open menu' }).click();

    // Click "Resolve"
    await page.getByRole('menuitem', { name: 'Resolve' }).click();

    // Verify status changes to "resolved"
    await expect(row.getByText('resolved')).toBeVisible();
  });
});
