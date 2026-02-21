/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices } from './e2e/test-data';

const HEADERS = { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' };

test.describe('Onboarding Flow', () => {
  test.beforeEach(async ({ request }) => {
    // Ensure a clean state before each test by deleting ALL services
    const res = await request.get('/api/v1/services', { headers: HEADERS });
    if (res.ok()) {
        const json = await res.json();
        const services = Array.isArray(json) ? json : (json.services || []);
        for (const s of services) {
            await request.delete(`/api/v1/services/${s.name}`, { headers: HEADERS });
        }
    }
  });

  test('should show onboarding hero when no services exist', async ({ page }) => {
    await page.goto('/');

    // Verify Hero Content
    await expect(page.getByRole('heading', { name: 'Welcome to MCP Any' })).toBeVisible();
    await expect(page.getByText('Your unified control plane')).toBeVisible();

    // Verify Actions
    await expect(page.getByRole('link', { name: 'Connect Service' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Browse Marketplace' })).toBeVisible();

    // Verify Dashboard Grid is NOT visible (Dashboard H1)
    await expect(page.getByRole('heading', { name: 'Dashboard' })).not.toBeVisible();
  });

  test('should show dashboard when services exist', async ({ page, request }) => {
    // Seed data
    await seedServices(request);

    await page.goto('/');

    // Verify Dashboard Header
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

    // Verify Hero is NOT visible
    await expect(page.getByRole('heading', { name: 'Welcome to MCP Any' })).not.toBeVisible();
  });
});
