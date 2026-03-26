import { test, expect } from '@playwright/test';

test.describe('Audit Logs Real Data View', () => {
  test('fetches real data and displays it using RichResultViewer', async ({ page, request }) => {
    // Seed the database with actual data via API
    const response = await request.post('/api/v1/debug/traces', {
      data: {
        action: 'seed_audit_logs',
        count: 5
      }
    });
    expect(response.ok()).toBeTruthy();

    // Navigate to the audit logs page
    await page.goto('/traces');

    // Wait for the table view to appear (indicates RichResultViewer detected an array)
    await expect(page.locator('table')).toBeVisible({ timeout: 10000 });

    // Verify tabular data formatting
    const rows = await page.locator('table tbody tr').count();
    expect(rows).toBeGreaterThan(0);

    // Verify specific column headers exist (proving it parsed the JSON)
    await expect(page.locator('th').filter({ hasText: 'EventType' })).toBeVisible();
    await expect(page.locator('th').filter({ hasText: 'Action' })).toBeVisible();

    // Verify raw JSON view is also present and populated
    await expect(page.locator('.react-json-view')).toBeVisible();
  });
});
