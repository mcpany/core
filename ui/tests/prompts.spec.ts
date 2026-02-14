import { test, expect } from '@playwright/test';

test('Prompt Workbench CRUD and Execution', async ({ page }) => {
  // Navigate to Prompts
  await page.goto('/prompts');
  await expect(page.getByText('Prompt Library')).toBeVisible();

  // Create New Prompt
  await page.getByTitle('Create New Prompt').click();
  await expect(page.getByRole('heading', { name: 'Create Prompt' })).toBeVisible();

  const timestamp = Date.now();
  const promptName = `e2e_test_${timestamp}`;
  const promptDesc = 'This is a test prompt';

  await page.getByLabel('Name').fill(promptName);
  await page.getByLabel('Description').fill(promptDesc);

  // Save
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText('Success')).toBeVisible();

  // Verify in list
  await expect(page.getByRole('button', { name: promptName })).toBeVisible();

  // Switch to Run tab
  await page.getByRole('tab', { name: 'Run / Preview' }).click();

  // Verify Runner
  await expect(page.getByText('Run: ' + promptName)).toBeVisible();

  // Fill argument (default schema has 'name')
  // We need to wait for arguments to appear
  await expect(page.getByLabel('name', { exact: true })).toBeVisible();
  await page.getByLabel('name', { exact: true }).fill('World');

  // Generate
  await page.getByRole('button', { name: 'Generate' }).click();

  // Verify output (default template is "Hello {{name}}, how are you?")
  await expect(page.getByText('Hello World, how are you?')).toBeVisible();

  // Edit again
  await page.getByRole('tab', { name: 'Edit Definition' }).click();
  await page.getByLabel('Description').fill(promptDesc + ' Updated');
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText('Success')).toBeVisible();

  // Delete
  page.on('dialog', dialog => dialog.accept());
  await page.getByRole('button', { name: 'Delete' }).click();
  await expect(page.getByText('Deleted')).toBeVisible();
  await expect(page.getByRole('button', { name: promptName })).not.toBeVisible();
});
