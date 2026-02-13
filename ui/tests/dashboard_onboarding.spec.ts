/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './e2e/test-data';

const HEADERS = { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' };

test.describe('Dashboard Onboarding', () => {
  // Use serial mode to ensure clean state transitions
  test.describe.configure({ mode: 'serial' });

  test.beforeAll(async ({ request }) => {
      await seedUser(request, "onboard-admin");
  });

  test.afterAll(async ({ request }) => {
      await cleanupUser(request, "onboard-admin");
      // Cleanup service created in test
      await request.delete('/api/v1/services/Onboarding Service', { headers: HEADERS }).catch(() => {});
  });

  test.beforeEach(async ({ page }) => {
      await page.goto('/login');
      // Wait for page to be fully loaded as it might be transitioning
      await page.waitForLoadState('networkidle');

      await page.fill('input[name="username"]', 'onboard-admin');
      await page.fill('input[name="password"]', 'password');
      await page.click('button[type="submit"]');

      // Wait for redirect to home page and verify
      await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test('Shows onboarding guide when empty', async ({ page, request }) => {
      // Ensure no services
      const res = await request.get('/api/v1/services', { headers: HEADERS });
      const data = await res.json();
      const services = Array.isArray(data) ? data : (data.services || []);

      for (const s of services) {
          await request.delete(`/api/v1/services/${s.id || s.name}`, { headers: HEADERS });
      }

      await page.reload();
      await expect(page.getByText('Welcome to MCP Any')).toBeVisible();
      await expect(page.getByText('1. Connect a Service')).toBeVisible();
  });

  test('Shows dashboard after adding service', async ({ page, request }) => {
      // Add a service
      const svc = {
          id: "svc_onboard",
          name: "Onboarding Service",
          version: "v1.0",
          http_service: {
              address: "http://example.com",
              tools: [],
              calls: {}
          }
      };
      const res = await request.post('/api/v1/services', { data: svc, headers: HEADERS });
      expect(res.ok()).toBeTruthy();

      await page.reload();
      await expect(page.getByText('Welcome to MCP Any')).not.toBeVisible();
      // Check for Network Graph visibility (it might take a moment for topology to update)
      // We can check for "Quick Actions" which is part of the standard grid
      await expect(page.getByText('Quick Actions')).toBeVisible();
  });

  test('Populates charts with traffic', async ({ page, request }) => {
      // Seed traffic
      // TrafficPoint format: { time: "HH:MM", requests: number, ... }
      const points = [
          { time: "12:00", requests: 50, errors: 0, latency: 100 },
          { time: "12:01", requests: 75, errors: 2, latency: 120 }
      ];
      const res = await request.post('/api/v1/debug/seed_traffic', { data: points, headers: HEADERS });
      expect(res.ok()).toBeTruthy();

      await page.reload();
      // Check that empty state message is gone
      await expect(page.getByText('No traffic data yet')).not.toBeVisible();
  });
});
