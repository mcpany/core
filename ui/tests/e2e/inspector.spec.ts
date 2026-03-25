import { test, expect } from '@playwright/test';
import { mockTrace } from './test-data';

test.describe('Inspector E2E', () => {
  test.beforeEach(async ({ request }) => {
    // Seed the database with our mock trace
    const seedResponse = await request.post('http://localhost:8080/api/debug/seed/traces', {
      data: [mockTrace]
    });
    expect(seedResponse.ok()).toBeTruthy();
  });

  test('should display table data using RichResultViewer in Trace Details', async ({ page }) => {
    await page.goto('/inspector');

    // Wait for the traces table to load
    await page.waitForSelector('text=call_tool');

    // Click on the trace row to open details
    await page.click('text=call_tool');

    // Expand the payload tab which now contains both request and response in our ui changes
    await page.click('button[role="tab"]:has-text("Payload")');

    // Verify the Input table renders correctly
    await expect(page.locator('text=Alice Smith')).toBeVisible();
    await expect(page.locator('text=alice@example.com')).toBeVisible();

    // Verify the Output table renders correctly
    await expect(page.locator('text=Users fetched successfully')).toBeVisible();
  });
});
