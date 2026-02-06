/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './e2e/test-data';

test.describe('Service Capabilities Badges', () => {
  const serviceId = "cap-service-test";
  const serviceName = "Capabilities Service";

  test.beforeEach(async ({ request, page }) => {
    // 1. Seed User
    await seedUser(request, "cap-admin");

    // 2. Seed Service with Capabilities
    // We define a service with 2 tools, 1 resource, and 3 prompts
    const serviceConfig = {
      id: serviceId,
      name: serviceName,
      version: "1.0",
      http_service: {
        address: "http://example.com",
        tools: [
          { name: "tool_a", description: "Tool A" },
          { name: "tool_b", description: "Tool B" }
        ],
        resources: [
          { uri: "res://a", name: "Resource A" }
        ],
        prompts: [
          { name: "prompt_a", description: "Prompt A" },
          { name: "prompt_b", description: "Prompt B" },
          { name: "prompt_c", description: "Prompt C" }
        ]
      }
    };

    const headers = { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' };
    const res = await request.post('/api/v1/services', { data: serviceConfig, headers });
    expect(res.ok()).toBeTruthy();

    // 3. Login
    await page.goto('/login');
    await page.fill('input[name="username"]', 'cap-admin');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test.afterEach(async ({ request }) => {
    const headers = { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' };
    await request.delete(`/api/v1/services/${serviceName}`, { headers });
    await cleanupUser(request, "cap-admin");
  });

  test('should display capability badges in service list', async ({ page }) => {
    await page.goto('/upstream-services');
    await expect(page.locator('h1')).toContainText('Upstream Services');

    // Find the row for our service
    const row = page.locator('tr', { hasText: serviceName });
    await expect(row).toBeVisible();

    // Check for Tools Badge (Zap icon + count 2)
    const toolsBadge = row.locator('.badge', { hasText: '2' }).filter({ has: page.locator('svg.lucide-zap') });
    // Note: Locator might be tricky with icon. Using title attribute if available or specific classes.
    // In code: title="Tools"
    await expect(row.locator('[title="Tools"]')).toContainText('2');

    // Check for Resources Badge (Database icon + count 1)
    await expect(row.locator('[title="Resources"]')).toContainText('1');

    // Check for Prompts Badge (MessageSquare icon + count 3)
    await expect(row.locator('[title="Prompts"]')).toContainText('3');
  });
});
