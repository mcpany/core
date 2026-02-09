import { test, expect } from '@playwright/test';

test('create and deploy a stack', async ({ page }) => {
  // 1. Navigate to Stacks page
  await page.goto('/stacks');

  // 2. Click "Add Stack"
  await page.click('text=Add Stack');
  await expect(page).toHaveURL(/\/stacks\/new/);

  // 3. Edit YAML
  // Focus the editor (Monaco uses a hidden textarea or contenteditable)
  // We can try to click the editor area and type.
  await page.click('.monaco-editor');

  // Clear existing content (cmd+a, backspace doesn't always work reliably in headless)
  // But our "new" template is small.
  // Let's just append or replace.
  // Actually, for reliability, let's use the clipboard or specific keyboard commands.
  // Or better, we can inject the value if we exposed a way, but we shouldn't modify app code for tests.

  // Let's try typing.
  // The default content is "# Define your stack here\nname: my-stack\nservices:\n"
  // We want to overwrite it.
  await page.keyboard.press('Control+A');
  await page.keyboard.press('Backspace');

  const yaml = `
name: e2e-stack
services:
  - name: e2e-time
    commandLineService:
      command: npx -y @modelcontextprotocol/server-time
`;
  await page.keyboard.insertText(yaml);

  // 4. Click Save
  await page.click('button:has-text("Save")');

  // 5. Verify redirection
  await expect(page).toHaveURL(/\/stacks\/e2e-stack/);

  // 6. Click Deploy
  // Wait for button to be enabled (might take a moment for validation state)
  await expect(page.locator('button:has-text("Deploy")')).toBeEnabled();
  await page.click('button:has-text("Deploy")');

  // 7. Verify Success Toast
  await expect(page.locator('text=Stack deployed successfully')).toBeVisible();

  // 8. Verify Service Created via UI
  await page.goto('/upstream-services');
  await expect(page.locator('text=e2e-time')).toBeVisible();
});
