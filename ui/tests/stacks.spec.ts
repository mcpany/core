import { test, expect } from '@playwright/test';

test.describe('Stack Management', () => {
  test('should create, edit, and list stacks', async ({ page }) => {
    // 1. Navigate to Stacks
    await page.goto('/stacks');
    await expect(page.locator('h1')).toContainText('Stacks');

    // 2. Click Create Stack
    await page.click('text=Create Stack');
    await expect(page).toHaveURL(/\/stacks\/new/);
    await expect(page.locator('h1')).toContainText('New Stack');

    // 3. Enter YAML content
    const stackName = `e2e-stack-${Date.now()}`;
    const yamlContent = `name: ${stackName}
description: E2E Test Stack
services:
  - name: weather-${Date.now()}
    commandLineService:
        command: npx -y @modelcontextprotocol/server-weather
`;

    // Wait for editor to load
    await page.waitForSelector('.monaco-editor');
    await page.click('.monaco-editor');

    // Clear editor content
    // Note: Control+A might vary by OS, but usually works for web editors in CI (Linux)
    await page.keyboard.press('Control+A');
    await page.keyboard.press('Backspace');

    // Type content
    await page.keyboard.type(yamlContent);

    // 4. Save
    await page.click('text=Deploy Stack');

    // 5. Verify redirection to list
    await expect(page).toHaveURL(/\/stacks$/);

    // 6. Verify stack in list
    await expect(page.locator(`text=${stackName}`)).toBeVisible();
  });
});
