import { test, expect } from '@playwright/test';

test('Live Trace Inspector and Replay Flow', async ({ page }) => {
  // Navigate to traces page
  await page.goto('http://localhost:9002/traces');

  // Verify Live Toggle exists
  const liveToggle = page.locator('button[title="Start Live Updates"]');
  await expect(liveToggle).toBeVisible();

  // Click Live Toggle
  await liveToggle.click();
  await expect(page.locator('button[title="Pause Live Updates"]')).toBeVisible();

  // Wait for and click on a trace (using the mock data which has "calculate_sum")
  const toolTrace = page.getByText('calculate_sum').first();
  await toolTrace.click();

  // Verify Replay button appears
  const replayButton = page.getByRole('button', { name: 'Replay in Playground' });
  await expect(replayButton).toBeVisible();

  // Click Replay and verify navigation
  await replayButton.click();
  await expect(page).toHaveURL(/\/playground\?tool=calculate_sum&args=/);

  // Verify Playground input
  const input = page.locator('input[placeholder*="e.g. calculator"]');
  await expect(input).toHaveValue(/calculate_sum/);
});
