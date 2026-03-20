import { test, expect } from '@playwright/test';

test.describe('Inspector Trace Delete', () => {
  test('should delete a seeded trace', async ({ page, request }) => {
    // Navigate to the inspector page
    await page.goto('/inspector');

    // Wait for the inspector to load initially
    await page.waitForLoadState('networkidle');

    // Seed a trace via the debug API
    // Need to use the correct API path since it's mounted on /api/v1
    const response = await request.post('/api/v1/traces/seed');
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    const traceId = data.id;

    // Refresh to ensure the trace appears
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Verify the trace is in the table (by matching the trace ID text somewhere)
    // The table might be long, so we look for the trace ID in the text
    await expect(page.locator(`text=${traceId}`)).toBeVisible();

    // The delete button is in the row with the trace id.
    // We can locate the row containing the trace ID and click the delete button within it.
    const row = page.locator('tr').filter({ hasText: traceId });
    const deleteButton = row.locator('button:has(.lucide-trash2)');
    await expect(deleteButton).toBeVisible();

    // Click to delete
    await deleteButton.click();

    // Verify toast appears
    await expect(page.locator('text=Trace deleted')).toBeVisible();

    // Verify the trace is removed from the DOM
    await expect(page.locator(`text=${traceId}`)).toBeHidden();
  });
});
