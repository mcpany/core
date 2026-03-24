import { test, expect } from '@playwright/test';

test.describe('Navigation Sidebar', () => {
  test('should display links to Approvals and Universal Agent Bus for admins', async ({ page }) => {
    // Navigate to the main page
    await page.goto('/');

    // Wait for the sidebar to load
    const sidebar = page.locator('nav'); // The exact tag might be different, let's wait for text

    // Switch to admin role using the dropdown if we aren't already an admin
    // Note: Based on the component logic, the default role might not be admin,
    // but the dropdown has a "Switch Role (Demo)" option.
    // Let's assume we are an admin or we can see it. If not, the test will fail and we can refine.
    // Actually, looking at app-sidebar.tsx, they are under the Platform group which is shown to admins.

    await expect(page.locator('text=Approvals')).toBeVisible();
    await expect(page.locator('text=Universal Agent Bus')).toBeVisible();

    // Click Approvals
    await page.locator('text=Approvals').click();
    await expect(page.locator('h2:has-text("HITL Approvals")')).toBeVisible();

    // Click Universal Agent Bus
    await page.locator('text=Universal Agent Bus').click();
    await expect(page.locator('h2:has-text("Universal Agent Bus")')).toBeVisible();
  });
});
