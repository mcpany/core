/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test('Network page loads and displays graph', async ({ page }) => {
  await page.goto('http://localhost:9002/network');

  // Check title
  await expect(page).toHaveTitle(/MCPAny Manager/);

  // Check for graph components
  await expect(page.locator('text=Network Topology')).toBeVisible();
  // Check for graph components
  await expect(page.locator('text=Network Topology')).toBeVisible();

  // Use a more specific selector for the node to avoid ambiguity or visibility issues
  const coreNode = page.locator('.react-flow__node').filter({ hasText: 'MCP Any' });
  await expect(coreNode).toBeVisible();

  // Check interaction
  await coreNode.click();
  await expect(page.locator('text=ID: mcp-core')).toBeVisible();
});
