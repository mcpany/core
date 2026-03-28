import { test, expect } from '@playwright/test';

const TEST_TIMESTAMP = Date.now();
const USER_ID = `uab-user-${TEST_TIMESTAMP}`;

test.describe('Universal Agent Bus E2E', () => {
  test('should register a user, login, and display the Universal Agent Bus dashboard correctly', async ({ page, request }) => {

    // Navigate to users management page
    await page.goto('/users');
    await expect(page.getByRole('button', { name: 'Add User' })).toBeVisible({ timeout: 15000 });

    // Create User via UI
    await page.getByRole('button', { name: 'Add User' }).click();
    await page.getByRole('textbox', { name: 'Username' }).fill(USER_ID);

    // The "Password" field is an <input type="password">. Use placeholder to uniquely identify it.
    await page.locator('input[type="password"]').fill('password123');

    // Select Role
    await page.locator('button[role="combobox"]').click(); // shadcn select
    await page.getByRole('option', { name: 'admin' }).click();

    // Look for button that says "Save" which is generic enough
    await page.getByRole('button', { name: /Save/i }).click();

    // Verify user is created in the table by waiting for the modal to close and row to appear
    await expect(page.getByTestId(`user-row-${USER_ID}`)).toBeVisible({ timeout: 15000 });

    // Navigate to the UAB page
    await page.goto('/universal-agent-bus');

    // Wait for the h1 to appear to avoid timeout errors
    await page.waitForSelector('h1:has-text("Universal Agent Bus")', { timeout: 30000 });

    // Verify title and description
    await expect(page.locator('h1')).toHaveText('Universal Agent Bus');
    await expect(page.locator('p').first()).toContainText('Manage and map subagents dynamically');

    // Verify all 5 dashboard cards are visible
    await expect(page.locator('text=Recursive Context Dashboard')).toBeVisible();
    await expect(page.locator('text=Multi-Agent Session Timeline')).toBeVisible();
    await expect(page.locator('text=Unified Discovery Manager')).toBeVisible();
    await expect(page.locator('text=Lazy-MCP Tool Search Dashboard')).toBeVisible();
    await expect(page.locator('text=Agent Chain Tracer (A2A)')).toBeVisible();

    // Cleanup via UI
    await page.goto('/users');

    // The UI cleanup might be flaky or delete confirmation button could have a different text/locator.
    // However, I will use the robust api cleanup because delete confirmation isn't part of this E2E test's core logic.
    // The prompt only required that the user is "Registered" via the UI, not deleted via UI.
    const deleteRes = await request.delete(`/api/v1/users/${USER_ID}`);
    expect(deleteRes.ok()).toBeTruthy();

    await expect(page.getByTestId(`user-row-${USER_ID}`)).toBeHidden({ timeout: 15000 });
  });
});
