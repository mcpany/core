import { test, expect } from '@playwright/test';

test.describe('Universal Agent Bus Interop E2E', () => {
  test('executes an interop task and verifies the adapter hub state without DB seeding', async ({ page, request }) => {
    // Navigate to UI and set token to bypass login as master API admin
    await page.goto('/');
    await page.evaluate((key) => localStorage.setItem('mcp_auth_token', 'Bearer ' + key), process.env.MCPANY_API_KEY || 'test-token');

    // Now create the user via UI
    await page.goto('/users');
    await expect(page.locator('h2:has-text("Users")')).toBeVisible({ timeout: 15000 });

    await page.click('button:has-text("Add User")');
    await expect(page.locator('h2:has-text("Add New User")')).toBeVisible();
    await page.fill('input[name="id"]', 'interop-tester');

    // The tabs might default to password, but let's make sure
    await page.click('button[role="tab"]:has-text("Password")');
    await page.fill('input[name="password"]', 'password123'); // Minimum 8 chars
    await page.click('button:has-text("Save Changes")');

    await expect(page.locator('div[role="dialog"]')).not.toBeVisible();

    // Now logout
    await page.evaluate(() => localStorage.removeItem('mcp_auth_token'));

    // Now login as the new user!
    await page.goto('/login');
    await page.fill('input[name="username"]', 'interop-tester');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL('/', { timeout: 15000 });

    // Navigate to the interop UI
    await page.goto('/universal-agent-bus');

    // Wait for the adapter list to load
    await expect(page.getByText('Registered Hub Adapters')).toBeVisible();
    await expect(page.getByText('OpenClaw, CrewAI, AutoGen')).toBeVisible();

    // The user inputs a task
    const input = page.getByPlaceholder('Intent (e.g. adaptive_reasoning)');
    await input.fill('adaptive_reasoning');

    // Execute task
    const executeBtn = page.getByRole('button', { name: 'Execute Task' });
    await executeBtn.click();

    // Assert that the result JSON block renders the success status
    await expect(page.getByText('"status": "success"')).toBeVisible({ timeout: 10000 });
  });
});
