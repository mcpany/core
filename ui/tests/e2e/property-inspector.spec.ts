import { test, expect } from '@playwright/test';

// Mock trace matching the shape that generateMockTrace() produces on the backend.
const MOCK_TRACE = {
  id: 'trace-seed-inspector-test',
  timestamp: new Date().toISOString(),
  totalDuration: 1250,
  status: 'success',
  trigger: 'user',
  rootSpan: {
    id: 'span-orchestrator-1',
    name: 'code-refactor',
    type: 'tool',
    status: 'success',
    startTime: 1000,
    endTime: 1150,
    input: { "file": "main.py", "action": "optimize" },
    output: {
        "diff":   "--- a/main.py\n+++ b/main.py\n@@ -1,5 +1,5 @@\n-def slow_func():\n-    pass\n+def fast_func():\n+    return True\n",
        "status": "success",
        "metadata": {
            "lines_changed": 4,
            "is_dry_run":    false,
            "warnings":      null,
            "tags":          ["optimization", "performance"],
            "ast_nodes": {
                "functions": 2,
                "classes":   0,
                "imports": [
                    {"module": "sys", "used": true},
                    {"module": "os", "used": false},
                ],
            },
        },
    },
    children: [],
  },
};

test.describe('Property Inspector', () => {
    test('verifies interactive JsonTree renders correctly for complex data', async ({ page }) => {
        // Intercept the POST request to /api/v1/debug/traces and simulate backend response
        await page.route('**/api/v1/debug/traces', async (route) => {
            await route.fulfill({
                status: 201,
                contentType: 'application/json',
                body: JSON.stringify({ status: 'seeded', id: MOCK_TRACE.id }),
            });
        });

        let wsSend: any = null;
        await page.routeWebSocket('**/api/v1/ws/traces', (ws: any) => {
            wsSend = (data: string) => ws.send(data);
        });

        // 1. Navigate to Inspector
        await page.goto('/inspector');

        // 2. Click Seed Trace
        const seedButton = page.getByRole('button', { name: /Seed Trace/i });
        await seedButton.waitFor({ state: 'visible', timeout: 30000 });
        await seedButton.click();

        if (wsSend) {
            wsSend(JSON.stringify(MOCK_TRACE));
        }

        // Wait for the trace to appear in the table.
        const traceRow = page.getByRole('row').filter({ hasText: 'code-refactor' }).first();
        await expect(traceRow).toBeVisible({ timeout: 10000 });

        // 3. Open the trace details
        await traceRow.click();

        // 4. Wait for the detail sheet to open
        const sheet = page.getByRole('dialog');
        await expect(sheet).toBeVisible();

        // 5. Navigate to Payload tab
        await sheet.getByRole('tab', { name: /Payload/i }).click();

        // 6. Verify the JsonTree (Property Inspector) is rendered instead of a raw SyntaxHighlighter
        // We look for the "Apple Design Standard" specific classes or texts we added
        const jsonTreeContainer = sheet.locator('.bg-muted\\/10.backdrop-blur-sm').first();
        await expect(jsonTreeContainer).toBeVisible();

        // 7. Expand the root object if not expanded
        // Click the first chevron in the response payload JsonTree to expand
        // It should contain the "metadata" key that we added in our backend seed
        const responsePayloadTree = sheet.locator('.bg-muted\\/10.backdrop-blur-sm').nth(1);
        await expect(responsePayloadTree).toBeVisible();

        // Verify the "metadata" key exists (we need to ensure it's expanded or we can just see it)
        await expect(responsePayloadTree.getByText('metadata', { exact: true })).toBeVisible();

        // Verify Type Badges exist (e.g. string, number, boolean)
        await expect(responsePayloadTree.getByText('string', { exact: true }).first()).toBeVisible();
        await expect(responsePayloadTree.getByText('number', { exact: true }).first()).toBeVisible();
    });
});