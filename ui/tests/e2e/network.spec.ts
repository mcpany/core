
import { test, expect } from '@playwright/test';

test('Network Graph loads and interacts', async ({ page }) => {
  // 1. Go to the Network page
  await page.goto('/network');

  // 2. Verify the graph container exists
  const graphContainer = page.locator('.react-flow');
  await expect(graphContainer).toBeVisible();

  // 3. Verify nodes are present
  // We expect "MCP Any Server" (Core) to be visible
  await expect(page.getByText('MCP Any Server')).toBeVisible();
  // We expect "Postgres DB" (Service) to be visible
  await expect(page.getByText('Postgres DB')).toBeVisible();

  // 4. Interact: Click on "Postgres DB" node
  await page.getByText('Postgres DB').click();

  // 5. Verify Side Sheet opens
  const sheetTitle = page.getByRole('heading', { name: 'Postgres DB' });
  await expect(sheetTitle).toBeVisible();

  // 6. Verify Metrics in Sheet
  await expect(page.locator('.text-lg.font-bold').first()).toBeVisible(); // Check for any metric value
  await expect(page.getByText('Latency')).toBeVisible();

  // 7. Take Screenshot
  await page.screenshot({ path: 'verification/network_graph_verified.png' });
});
