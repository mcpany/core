/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('dashboard network topology widget', async ({ page }) => {
  // Go to dashboard
  await page.goto('/');

  // The widget should be present in the default layout.
  // We check for the React Flow renderer container which is more stable than .react-flow class in some versions
  // Or check for the parent container we defined in the widget
  // The widget is wrapped in a border box

  // Wait for React Flow to initialize
  await expect(page.locator('.react-flow__renderer').or(page.locator('.react-flow'))).toBeVisible({ timeout: 30000 });

  // Check for the presence of nodes.
  // React Flow nodes usually have the class 'react-flow__node'
  // We wait for at least one node to appear (it might take a moment to fetch topology)
  await expect(page.locator('.react-flow__node').first()).toBeVisible({ timeout: 15000 });

  // Optional: Check if "Core" node is present (assuming seeded data has Core)
  // The label might be inside the node.
  // Note: The specific label depends on the backend seeding.
  // But we at least verified the graph loaded and rendered nodes.
});
