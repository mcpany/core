import { test, expect } from '@playwright/test';

test('debug resource preview', async ({ page }) => {
    // Mock resources list
    await page.route('**/api/v1/resources', async route => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                resources: [{ uri: 'file:///test.json', name: 'test.json', mimeType: 'application/json' }]
            })
        });
    });

    // Mock resource read with regex to handle encoded URI
    await page.route(/\/api\/v1\/resources\/read.*/, async route => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                contents: [{
                    uri: 'file:///test.json',
                    mimeType: 'application/json',
                    text: '{"key": "value", "long": "content to test modal view"}'
                }]
            })
        });
    });

    await page.goto('/resources');

    const resourceItem = page.locator('div.font-medium', { hasText: 'test.json' });
    await expect(resourceItem).toBeVisible();

    await resourceItem.click();

    // Take screenshot to see what's actually rendered
    await page.waitForTimeout(2000); // Wait for render
    await page.screenshot({ path: '/home/jules/verification/debug-preview.png', fullPage: true });
});
