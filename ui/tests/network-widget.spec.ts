/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('dashboard network topology widget', async ({ page }) => {
  // Go to dashboard
  await page.goto('/');

  // The widget should be present in the default layout.
  // We can look for the React Flow container
  await expect(page.locator('.react-flow')).toBeVisible();

  // We can also check if the widget title "Network Topology" is visible in the layout configuration (if we opened it)
  // But we want to check the widget itself.

  // Check for the presence of nodes.
  // React Flow nodes usually have the class 'react-flow__node'
  // We wait for at least one node to appear (it might take a moment to fetch topology)
  await expect(page.locator('.react-flow__node').first()).toBeVisible({ timeout: 10000 });

  // Optional: Check if "Core" node is present (assuming seeded data has Core)
  // The label might be inside the node.
  // Note: The specific label depends on the backend seeding.
  // But we at least verified the graph loaded and rendered nodes.
});
