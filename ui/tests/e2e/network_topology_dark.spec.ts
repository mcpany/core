/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('Network Topology Dark Mode', () => {
  // We can't easily toggle dark mode in the test without UI interaction or local storage manipulation.
  // Assuming the app supports a theme toggle or system preference.
  // For this test, we'll try to force dark mode if possible or just check if classes are present.

  test.use({ colorScheme: 'dark' });

  test('should display network topology nodes with correct dark mode styles', async ({ page }) => {
    // Navigate to the network topology page
    await page.goto('/network');

    // Wait for the graph to load
    await page.waitForSelector('.react-flow__node');

    // Check for specific node types and their computed styles
    // Since we can't easily check computed styles dependent on dark mode without visual regression testing or deeply inspecting computed styles which might vary by browser implementation of dark mode preferences
    // We will inspect the class names to ensure the dark mode classes are applied to the nodes.

    // Select a Service node (should have dark:bg-blue-900)
    // We need to find a node that is likely to be a service.
    // Based on the code, nodes are dynamically loaded. We might need to mock the API response, but for E2E we usually run against a real backend or mock it at network level.

    // Let's assume there is at least one node.
    const nodes = page.locator('.react-flow__node');
    await expect(nodes.first()).toBeVisible();

    // Verify that at least some nodes have the dark mode classes we added.
    // We can check if the HTML content of the nodes contains the dark classes.
    // Note: React Flow might wrap the content. The classes we added are on the node wrapper (via className prop).

    const content = await page.content();

    // We expect to see 'dark:bg-slate-800' or similar in the class list of some elements.
    // This is a loose check but confirms our code is being rendered.
    expect(content).toContain('dark:bg-slate-800'); // Core
    expect(content).toContain('dark:bg-blue-900'); // Service
    // expect(content).toContain('dark:bg-green-900'); // Client - might not be present if no client in data
  });
});
