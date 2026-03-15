import { test, expect } from '@playwright/test';

test.describe('Inspector E2E', () => {
  test('seeds a trace and views its details', async ({ page, request }) => {
    // 1. Seed a trace using the backend debug endpoint
    const seedResponse = await request.post('/api/v1/debug/traces', {
      headers: {
        'Content-Type': 'application/json',
      },
      data: {}
    });

    expect(seedResponse.ok()).toBeTruthy();
    const result = await seedResponse.json();
    const traceId = result.id;
    expect(traceId).toBeTruthy();

    // 2. Navigate to the Inspector page
    await page.goto('/inspector');

    // 3. Verify the trace appears in the table
    // The table should contain the trace ID or a row representing the trace
    const traceRow = page.locator('tr').filter({ hasText: traceId });
    await expect(traceRow).toBeVisible({ timeout: 10000 });

    // 4. Click the trace to open the TraceDetail sheet
    await traceRow.click();

    // Verify the TraceDetail sheet opens by checking for its content
    // We look for a heading or text specific to the trace details view
    const sheetHeader = page.locator('.sm\\:max-w-\\[800px\\]');
    await expect(sheetHeader).toBeVisible();

    // Check that the tokens section or some detail component is rendered inside
    const tokensLabel = page.getByText('Tokens', { exact: true });
    await expect(tokensLabel).toBeVisible();
  });
});
