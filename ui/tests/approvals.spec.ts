import { test, expect } from '@playwright/test';

test.describe('Approvals Page', () => {
  test('should load the HITL Approvals page successfully', async ({ page }) => {
    // We navigate to the approvals page directly
    await page.goto('/approvals');

    // Wait for the main heading to be visible
    const heading = page.locator('h2:has-text("HITL Approvals")');
    await expect(heading).toBeVisible();

    // Verify the page structure
    await expect(page.locator('text=Pending Approvals')).toBeVisible();
    await expect(page.locator('text=Approval Queue')).toBeVisible();
    await expect(page.locator('text=No pending approvals found.')).toBeVisible();
  });
});
