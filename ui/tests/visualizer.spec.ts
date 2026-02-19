/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser } from './e2e/test-data';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY };

test.describe('Agent Flow Visualizer', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(async ({ request, page }) => {
    // Seed Services (to have Service Nodes)
    await seedServices(request);
    await seedUser(request, "viz-admin");

    // Login
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'viz-admin');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]', { force: true });
    await page.waitForURL('/', { timeout: 30000 });
  });

  test.afterEach(async ({ request }) => {
    await cleanupServices(request);
  });

  test('Visualizer renders real topology data', async ({ page, request }) => {
    // 1. Seed Traffic to generate "Client" nodes
    // The backend expects "HH:MM" format for time
    const now = new Date();
    const timeStr = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;

    const trafficPoints = [
        {
            time: timeStr,
            requests: 50,
            errors: 0,
            latency: 20
        }
    ];

    const res = await request.post(`${BASE_URL}/api/v1/debug/seed_traffic`, {
        data: trafficPoints,
        headers: HEADERS
    });
    expect(res.ok()).toBeTruthy();

    // 2. Navigate to Visualizer
    await page.goto('/visualizer');

    // 3. Assertions
    // Check if React Flow is loaded
    await expect(page.locator('.react-flow')).toBeVisible();

    // Check for "MCP Core" node (The central node)
    // The label is "MCP Any" in backend manager.go: coreNode label "MCP Any"
    // We scope to .react-flow to avoid matching the header text "MCP Any"
    await expect(page.locator('.react-flow').getByText('MCP Any')).toBeVisible({ timeout: 10000 });

    // Check for Service Nodes (seeded by seedServices)
    // "Payment Gateway"
    await expect(page.locator('text=Payment Gateway')).toBeVisible();

    // Check for Client Node (seeded by seed_traffic)
    // The backend creates a session with ID "seed-data"
    await expect(page.locator('text=seed-data')).toBeVisible();

    // Check for Tool Nodes (children of Payment Gateway)
    // "process_payment"
    // Since we expanded children in the hook, it should be visible or connected
    await expect(page.locator('text=process_payment')).toBeVisible();
  });
});
