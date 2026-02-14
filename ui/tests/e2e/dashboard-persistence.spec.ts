import { test, expect } from '@playwright/test';

test.describe('Dashboard Persistence', () => {
  test('should persist widget layout to backend', async ({ page, request }) => {
    // 1. Seed the DB with a known layout
    const initialLayout = [
      {
        instanceId: 'test-widget-1',
        type: 'recent-activity',
        title: 'Test Activity',
        size: 'third',
        hidden: false
      }
    ];

    const seedRes = await request.post('/api/v1/user/preferences', {
      data: {
        'dashboard-layout': JSON.stringify(initialLayout)
      }
    });
    expect(seedRes.ok()).toBeTruthy();

    // 2. Visit Dashboard
    await page.goto('/');

    // 3. Verify the widget is loaded from backend
    await expect(page.getByText('Test Activity')).toBeVisible();

    // 4. Modify the layout (e.g. hide the widget)
    // We can use the "Layout" button to hide it
    await page.getByRole('button', { name: 'Layout' }).click();
    await page.locator('label').filter({ hasText: 'Test Activity' }).click(); // Toggle checkbox

    // 5. Wait for debounce save (500ms) + network latency
    await page.waitForTimeout(2000);

    // 6. Reload page
    await page.reload();

    // 7. Verify widget is NOT visible (hidden persisted)
    await expect(page.getByText('Test Activity')).not.toBeVisible();

    // 8. Verify backend state directly
    const backendRes = await request.get('/api/v1/user/preferences');
    expect(backendRes.ok()).toBeTruthy();
    const prefs = await backendRes.json();
    const savedLayout = JSON.parse(prefs['dashboard-layout']);

    expect(savedLayout).toHaveLength(1);
    expect(savedLayout[0].hidden).toBe(true);
  });
});
