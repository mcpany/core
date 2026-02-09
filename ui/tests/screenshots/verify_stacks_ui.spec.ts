
import { test, expect } from '@playwright/test';
import { seedCollection, cleanupCollection } from '../e2e/test-data';

test('verify stacks ui', async ({ page, request }) => {
    const stackName = 'screenshot-stack';
    await cleanupCollection(stackName, request);
    await seedCollection(stackName, request);

    await page.goto('/stacks');
    await expect(page.getByText(stackName)).toBeVisible();

    // Screenshot list
    await page.screenshot({ path: 'tests/screenshots/stacks-page.png', fullPage: true });

    // Open dialog
    await page.getByRole('button', { name: 'Create Stack' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    // Screenshot dialog
    await page.screenshot({ path: 'tests/screenshots/create-stack-dialog.png' });

    await cleanupCollection(stackName, request);
});
