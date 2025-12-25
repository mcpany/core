
import { test, expect } from '@playwright/test';

test('Network Graph loads and renders nodes', async ({ page }) => {
  // Mock the API response
  await page.route('/api/services', async route => {
    const json = [
      {
        id: "svc_test_01",
        name: "Test Payment Gateway",
        version: "v1.0.0",
        disable: false,
        service_config: { http_service: { address: "https://api.stripe.com" } }
      }
    ];
    await route.fulfill({ json });
  });

  await page.goto('/graph');

  // Verify Page Title
  await expect(page.getByText('Network Graph')).toBeVisible();
  await expect(page.getByText('Visualize the topology')).toBeVisible();

  // Verify Nodes
  // Note: We use .locator to find elements rendered by ReactFlow.
  // Depending on how ReactFlow renders, we might need specific selectors.
  // The Nodes have "Test Payment Gateway" text inside them.
  await expect(page.getByText('Test Payment Gateway')).toBeVisible();
  await expect(page.getByText('MCP Any')).toBeVisible();

  // Verify Status Badge
  await expect(page.getByText('Online')).toBeVisible();
});
