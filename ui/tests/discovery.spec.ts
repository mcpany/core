import { test, expect } from '@playwright/test';
import path from 'path';

test('Discovery Dashboard loads and displays providers', async ({ page }) => {
  // Navigate to upstream services page
  await page.goto('/upstream-services');

  // Verify we are on the page
  await expect(page.getByRole('heading', { name: 'Upstream Services' })).toBeVisible();

  // Click on the Discovery tab
  await page.getByRole('tab', { name: 'Auto-Discovery' }).click();

  // Verify Dashboard elements
  await expect(page.getByText('Auto-Discovery Providers')).toBeVisible();

  // Verify Ollama provider card is present. Use exact match for the name in the title.
  await expect(page.getByText('ollama', { exact: true })).toBeVisible();

  // Verify error message is also present (since we expect it to fail in dev/test)
  await expect(page.getByText('ollama not found', { exact: false })).toBeVisible();

  // Verify we can click Refresh
  const refreshButton = page.getByRole('button', { name: 'Refresh' });
  await expect(refreshButton).toBeVisible();
  await refreshButton.click();

  // Verify that the card is still visible after refresh
  await expect(page.getByText('ollama', { exact: true })).toBeVisible();

  // Take a screenshot
  await page.screenshot({ path: path.join(process.cwd(), '..', 'verification', 'discovery_dashboard.png') });
});
