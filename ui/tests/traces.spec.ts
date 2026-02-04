import { test, expect } from '@playwright/test';

test.describe('Traces Page', () => {
  test('should load traces with correct limit', async ({ page }) => {
    // Intercept the API call to verify query parameters
    let requestUrl: string | null = null;
    await page.route('**/api/traces*', async (route) => {
        requestUrl = route.request().url();
        await route.continue();
    });

    await page.goto('/traces');

    // Verify the page title or header
    await expect(page.getByText('Traces')).toBeVisible();

    // Verify the API was called with limit=100
    expect(requestUrl).toContain('limit=100');
  });

  test('should display trace details when clicked', async ({ page }) => {
      // Mock data for this test to ensure UI behavior
      await page.route('**/api/traces*', async (route) => {
          await route.fulfill({
              json: [
                  {
                      id: 'trace-1',
                      timestamp: new Date().toISOString(),
                      totalDuration: 123,
                      status: 'success',
                      trigger: 'user',
                      rootSpan: {
                          id: 'span-1',
                          name: 'test-tool',
                          serviceName: 'core',
                          startTime: Date.now(),
                          endTime: Date.now() + 123,
                          status: 'success',
                          type: 'tool'
                      }
                  }
              ]
          });
      });

      await page.goto('/traces');

      // Check if trace item is visible
      await expect(page.getByText('test-tool')).toBeVisible();

      // Click it
      await page.getByText('test-tool').click();

      // Verify detail view
      await expect(page.getByText('Execution Detail')).toBeVisible();
  });
});
