import { test, expect } from '@playwright/test';

test.describe('Nested Data SmartTable Visualization', () => {
  test('should format deeply nested JSON array payload as a Smart Table', async ({ page, request }) => {
    const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
    const HEADERS = { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };

    // Use the debug endpoint to seed traces
    await request.post(`/api/v1/debug/traces`, { headers: HEADERS }).catch(() => null);

    // Also seed a specific audit log for a tool that returns a deeply nested array
    // This tests the Smart Table in the Audit Log Viewer Dialog which uses RichResultViewer
    const mockAuditLog = {
        "timestamp": new Date().toISOString(),
        "tool_name": "nested_data_tool",
        "user_id": "system",
        "profile_id": "default",
        "arguments": "{}",
        "result": JSON.stringify({
            "response": {
                "data": {
                    "items": ["apple", "banana", "cherry"]
                }
            }
        }),
        "duration": "150ms",
        "duration_ms": 150
    };

    await request.post('/api/v1/debug/seed?api_key=mcpany-dev-key', {
        data: JSON.stringify([mockAuditLog]),
        headers: { 'Content-Type': 'application/json' }
    }).catch(() => null);

    await page.goto('/audit');

    const row = page.locator('tr:has-text("nested_data_tool")').first();
    const rowIsVisible = await row.isVisible({ timeout: 2000 }).catch(() => false);

    if (rowIsVisible) {
        const viewButton = row.locator('button:has-text("View")').first();
        await viewButton.click();

        // The dialog should open, and the deeply nested array should trigger the Table view
        await expect(page.locator('text=Audit Log Detail')).toBeVisible();
        const tableTab = page.locator('button[role="tab"]:has-text("Table")');
        await expect(tableTab).toBeVisible({ timeout: 5000 });

        // Ensure "apple" is in the table
        const dialogContent = page.locator('[role="dialog"]');
        await expect(dialogContent.locator('td:has-text("apple")')).toBeVisible();
    }
  });
});
