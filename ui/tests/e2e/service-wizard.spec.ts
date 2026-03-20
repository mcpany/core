import { test, expect } from '@playwright/test';

test.describe('Service Creation Wizard', () => {
  test('should create a new HTTP service via the wizard', async ({ page }) => {
    await page.goto('/upstream-services');

    // Open Wizard
    await page.click('button:has-text("Add Service")');

    // Step 1: Template Selection
    await expect(page.locator('text="Choose a Template"')).toBeVisible();
    await page.click('button:has-text("Start from Blank Service")');

    // Step 2: Basic Info
    await expect(page.locator('text="Service Details"')).toBeVisible();

    // Fill required fields
    await page.fill('input[id="name"]', 'Wizard E2E Test API');

    // Select HTTP service type
    await page.click('button:has-text("Not Configured")'); // Open select
    await page.click('text="HTTP / REST"'); // Select HTTP

    await page.fill('input[placeholder="https://api.example.com"]', 'https://echo.free.beeceptor.com');

    await page.click('button:has-text("Next")');

    // Step 3: Auth
    await expect(page.locator('text="Authentication"')).toBeVisible();
    await page.click('button:has-text("Next")'); // Skip auth for this test

    // Step 4: Review
    await expect(page.locator('text="Review & Create"')).toBeVisible();
    await expect(page.locator('text="Wizard E2E Test API"')).toBeVisible();
    await expect(page.locator('text="https://echo.free.beeceptor.com"')).toBeVisible();

    // Submit
    await page.click('button:has-text("Create Service")');

    // Wait for the dialog to close and success toast
    await expect(page.locator('text="Service Created"')).toBeVisible();

    // Verify it appears in the list (this tests real data is fetched)
    await expect(page.locator('text="Wizard E2E Test API"').first()).toBeVisible();
  });
});
