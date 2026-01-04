/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('network topology loads and displays nodes', async ({ page }) => {
  await page.goto('/network');

  // Wait for the graph to load
  await expect(page.getByText('Network Graph')).toBeVisible();
  await expect(page.getByText('Live topology of MCP services and tools.')).toBeVisible();

  // Check for presence of key nodes (from mock data)
  // React Flow renders nodes as divs with text
  await expect(page.getByText('MCP Any')).toBeVisible();
  await expect(page.getByText('weather-service')).toBeVisible();

  // Test interaction
  await page.getByText('weather-service').click();

  // Sheet should open
  await expect(page.getByText('Operational Status')).toBeVisible();
});
