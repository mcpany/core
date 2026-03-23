/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('E2E: Tool Status Update', async ({ page, request }) => {
  // Add a test service to ensure we have a tool to modify
  const svcConfig = {
    name: "Test Tools Service",
    id: "test-tools-svc",
    http_service: {
      address: "http://localhost:50050",
      tools: [
        {
          name: "test_tool_e2e",
          description: "Test tool for E2E",
          call_id: "test_tool_call"
        }
      ],
      calls: {
        test_tool_call: {
          method: "HTTP_METHOD_GET",
          endpoint_path: "/test"
        }
      }
    }
  };

  // Seed the database
  const createRes = await request.post('/api/v1/debug/seed_traffic', {
    data: {
      services: [svcConfig]
    },
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': process.env.MCPANY_API_KEY || 'test-token'
    }
  });

  expect(createRes.ok()).toBeTruthy();

  // Navigate to tools page
  await page.goto('/tools');

  // Verify the tool appears in the list
  await expect(page.getByText('test_tool_e2e')).toBeVisible();

  // Find the toggle button
  // Note: the ui toggles state locally immediately but relies on the PUT request to persist.
  // The toggle button is usually a Switch component.
  const row = page.locator('tr').filter({ hasText: 'test_tool_e2e' });
  const toggle = row.locator('button[role="switch"]');

  // Click the toggle to disable
  await toggle.click();

  // Wait for network request
  const response = await page.waitForResponse(response =>
    response.url().includes('/api/v1/tools') && response.request().method() === 'PUT'
  );

  expect(response.ok()).toBeTruthy();
  const resData = await response.json();
  expect(resData.name).toBe('test_tool_e2e');
  expect(resData.disable).toBe(true);

  // Reload page to verify persistence
  await page.goto('/tools');
  await expect(page.getByText('test_tool_e2e')).toBeVisible();

  const reloadedRow = page.locator('tr').filter({ hasText: 'test_tool_e2e' });
  const reloadedToggle = reloadedRow.locator('button[role="switch"]');

  // aria-checked should be "false" since we disabled it (assuming disabled = false in Switch)
  // Wait, let's check the UI logic:
  // <Switch checked={!tool.disable} onCheckedChange={(c) => toggleTool(tool.name, !c)} />
  // So aria-checked should be "false"
  await expect(reloadedToggle).toHaveAttribute('aria-checked', 'false');
});
