import { test, expect } from '@playwright/test';

test('Create Config Wizard with Registry Template', async ({ page }) => {
    // 1. Navigate to Marketplace
    await page.goto('/marketplace');

    // 2. Click Create Config
    const createBtn = page.getByRole('button', { name: 'Create Config' });
    await expect(createBtn).toBeVisible();
    await createBtn.click();

    // 3. Verify Service Type step
    await expect(page.locator('input#service-name')).toBeVisible();

    // 4. Select Template (PostgreSQL from registry)
    await page.getByRole('combobox', { name: 'Template' }).click();
    await page.getByRole('option', { name: 'PostgreSQL' }).first().click();

    // 5. Enter Name
    await page.fill('input#service-name', 'Test Postgres Config');

    // 6. Click Next
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 7. Verify Schema Form appears
    await expect(page.getByText('This service has a configuration schema')).toBeVisible();
    await expect(page.getByText('Connection URL')).toBeVisible();
});
