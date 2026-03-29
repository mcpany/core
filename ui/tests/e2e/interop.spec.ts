import { test, expect } from '@playwright/test';

test.describe('Interop API Integration', () => {

  test('should register, login, and run an interop task', async ({ page }) => {
    const username = `interop_${Date.now()}`;
    const password = 'Password123!';

    // Register
    await page.goto('/login');

    // Click the Sign Up / Create Account tab
    await page.locator('text=Create Account').or(page.locator('text=Sign Up')).click();

    // Fill Registration Form
    await page.getByPlaceholder('name@example.com').or(page.getByLabel('Email')).fill(`${username}@example.com`);
    await page.getByPlaceholder('••••••••').or(page.getByLabel('Password')).first().fill(password);
    await page.getByLabel('Confirm Password').or(page.locator('input[name="confirmPassword"]')).fill(password);

    await page.getByRole('button', { name: /Sign Up|Create Account/ }).click();
    await page.waitForURL('**/', { timeout: 10000 });

    // Go to Interop page
    await page.goto('/interop');

    // Verify elements
    await expect(page.locator('h2')).toHaveText('Interop Tester', { timeout: 10000 });

    // Fill form
    await page.locator('#framework-input').fill('CrewAI');
    await page.locator('#intent-input').fill('task_delegation');
    await page.locator('#payload-role-input').fill('data_analyst');

    // Submit
    await page.locator('#submit-interop-btn').click();

    // Verify Result
    const resultElement = page.locator('#interop-result');
    await expect(resultElement).toBeVisible({ timeout: 10000 });
    await expect(resultElement).toContainText('"status": "success"');
  });
});
