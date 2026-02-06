import { test, expect } from '@playwright/test';

test('OpenAPI Import Wizard Preview Flow', async ({ page }) => {
  // 1. Navigate to Services
  await page.goto('/upstream-services');
  await expect(page.getByRole('heading', { name: 'Upstream Services' })).toBeVisible();

  // 2. Click Add Service
  await page.getByRole('button', { name: 'Add Service' }).click();

  // 3. Select Custom Service template
  await page.getByText('Custom Service').click();

  // 4. Enter General Details
  await page.getByLabel('Service Name').fill('petstore-demo');

  // 5. Select OpenAPI Type
  await page.getByRole('tab', { name: 'Connection' }).click();

  // Combobox handling for Shadcn UI
  await page.getByLabel('Service Type').click();
  await page.getByRole('option', { name: 'OpenAPI / Swagger' }).click();

  // 6. Enter Connection Details
  await page.getByLabel('Base Address').fill('https://petstore.swagger.io/v2');
  await page.getByLabel('Spec URL').fill('https://petstore.swagger.io/v2/swagger.json');

  // 7. Click Preview
  const previewBtn = page.getByRole('button', { name: 'Preview & Import' });
  await expect(previewBtn).toBeVisible();
  await previewBtn.click();

  // 8. Verify Tools Discovered
  // Expect a toast or the list. Using a generous timeout for network fetch.
  await expect(page.getByText('Preview Successful')).toBeVisible({ timeout: 20000 });

  // Verify list appears
  await expect(page.getByText('Discovered Tools')).toBeVisible();
  // Verify specific tools from Petstore
  await expect(page.getByText('addPet')).toBeVisible();
  await expect(page.getByText('getPetById')).toBeVisible();

  // 9. Save
  await page.getByRole('button', { name: 'Save Changes' }).click();
  await expect(page.getByText('Service Created')).toBeVisible();

  // 10. Verify in List
  // Wait for dialog/sheet to close
  await expect(page.getByRole('heading', { name: 'Upstream Services' })).toBeVisible();
  await expect(page.getByText('petstore-demo')).toBeVisible();
});
