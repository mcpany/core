import { test, expect } from '@playwright/test';

test('Add Environment Variable to Command Line Service', async ({ page }) => {
  // 1. Seed Database (Indirectly via UI for now as per constraints, assuming empty state)
  // Actually, we can just start fresh.

  await page.goto('/');

  // 2. Open Register Service Dialog
  await page.getByRole('button', { name: 'Add Service' }).click(); // From Upstream Services page
  // Or "Register Service" if on Dashboard? Let's assume Services page or navigation
  // Let's navigate to Services page first if not default
  if (await page.getByRole('heading', { name: 'Upstream Services' }).count() === 0) {
      // Find nav
      // Assuming Sidebar
      await page.goto('/upstream-services');
  }

  // Wait for dialog trigger
  await page.getByRole('button', { name: 'Add Service' }).click();

  // 3. Select Command Line Type
  await page.getByText('Start from scratch').click(); // If template selector appears
  // Or if it goes straight to form:
  // We need to handle the template selector step if it exists.
  // Based on code: view starts as "templates".
  // So we click "Start from scratch" (which is usually a card or button in ServiceTemplateSelector)
  // Let's assume "Create from Scratch" or similar.
  // Inspecting ServiceTemplateSelector code would be ideal, but let's guess based on typical UI.
  // Actually, the previous code showed `ServiceTemplateSelector`.

  // Let's wait for "Basic Configuration" tab to ensure we are in form
  // If not, click the scratch option.
  if (await page.getByRole('tab', { name: 'Basic Configuration' }).count() === 0) {
      // Find the "scratch" template or button.
      // Usually the first card or a distinct button.
      // Let's look for "Empty" or "Scratch".
      await page.getByText('Empty Service').click();
  }

  // 4. Fill Basic Form
  await page.getByLabel('Service Name').fill('test-env-service');

  // Select Type: Command Line
  await page.getByLabel('Service Type').click();
  await page.getByRole('option', { name: 'Command Line' }).click();

  // Fill Command
  await page.getByLabel('Command').fill('echo $TEST_VAR');

  // 5. Add Environment Variable
  await page.getByRole('button', { name: 'Add Variable' }).click();
  await page.getByPlaceholder('KEY').fill('TEST_VAR');
  await page.getByPlaceholder('VALUE').fill('hello_world');

  // 6. Verify JSON Sync (Optional but good)
  await page.getByRole('tab', { name: 'Advanced (JSON)' }).click();
  await expect(page.locator('textarea')).toContainText('"TEST_VAR": "hello_world"');

  // 7. Submit
  await page.getByRole('button', { name: 'Register Service' }).click();

  // 8. Verify Success Toast
  await expect(page.getByText('Service Registered')).toBeVisible();

  // 9. Verify in List
  await expect(page.getByText('test-env-service')).toBeVisible();
});
