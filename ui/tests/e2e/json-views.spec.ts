import { test, expect, request } from '@playwright/test';

const SERVICE_ID = 'jsonview-test-service';

const SERVICE_CONFIG = {
  name: SERVICE_ID,
  id: SERVICE_ID,
  version: "1.0.0",
  command_line_service: {
    command: "echo",
    tools: [
      {
        "name": "test_json_schema",
        "call_id": "test_json_schema",
        "description": "Test tool with complex schema",
        "input_schema": {
          "type": "object",
          "properties": {
            "configData": { "type": "string", "description": "The config data" },
            "nested": {
              "type": "object",
              "properties": {
                "key1": { "type": "string" },
                "key2": { "type": "integer" }
              }
            }
          }
        }
      }
    ],
    calls: {
      "test_json_schema": {
        "args": ["-e", "echo '{\"status\": \"success\"}'"]
      }
    }
  }
};

test.describe('JsonView UI Components', () => {
  let apiContext: any;

  test.beforeAll(async () => {
    apiContext = await request.newContext({
      baseURL: process.env.BACKEND_URL || 'http://localhost:50050',
      extraHTTPHeaders: {
        'Content-Type': 'application/json',
      }
    });

    // Seed the database with our test service
    await apiContext.post('/api/v1/services', {
      data: SERVICE_CONFIG
    });
  });

  test.afterAll(async () => {
    // Clean up database
    await apiContext.delete(`/api/v1/services/${SERVICE_ID}`);
    await apiContext.dispose();
  });

  test('should verify ToolRunner renders JsonView for schema rather than raw string', async ({ page }) => {
    await page.goto('/tools');

    // Wait for table to render rows, we might need to wait for API call to fetch tools
    await page.waitForResponse('**/api/v1/tools*');

    // Wait for the specific tool to appear in the table
    // Sometimes the server is a bit slow responding with the updated tool
    // We add a reload loop to make sure it picks up the newly registered tool from beforeAll
    for (let i = 0; i < 10; i++) {
        try {
            await expect(page.getByText(/test_json_schema/i).first()).toBeVisible({ timeout: 5000 });
            break;
        } catch (e) {
            if (i === 9) throw e;
            await page.reload();
        }
    }

    // Filter to ensure test tool is found
    const searchInput = page.getByPlaceholder('Search tools...');
    await searchInput.fill('test_json_schema');
    await page.waitForTimeout(500); // small wait for search filter

    // Open the inspector for our test tool
    // We use .first() or .nth(0) in case there are multiple matching rows, but we only expect one
    const row = page.getByRole('row').filter({ hasText: /test_json_schema/i }).first();
    const inspectBtn = row.getByRole('button', { name: /Inspect/i }).first();

    // Ensure button is ready to receive clicks
    await expect(inspectBtn).toBeEnabled({ timeout: 10000 });

    // Use force true if there's a minor overlay
    await inspectBtn.click({ force: true });

    // Verify dialog opens
    await expect(page.getByRole('heading', { name: 'test_json_schema' })).toBeVisible({ timeout: 10000 });

    // Switch to Schema tab
    await page.getByRole('tab', { name: 'Schema' }).click();

    // Switch to JSON sub-tab
    await page.getByRole('tab', { name: 'JSON' }).click();

    // Verify JsonView is rendered instead of a raw `<pre>` string.
    // JsonView renders standard tree view items. For example, the primitive string keys will have quotes but it's interactive.
    // The class 'bg-[#1e1e1e]' is used for JsonView container.
    const jsonViewContainer = page.locator('.bg-\\[\\#1e1e1e\\]').first();
    await expect(jsonViewContainer).toBeVisible();

    // In JsonView tree mode, we can expand/collapse. Look for the "object" signature.
    // Ensure "configData" text is visible inside the JsonView (it should have quotes around the key)
    await expect(page.getByText('"configData":').first()).toBeVisible();

    // Check that we can collapse/expand which proves it's interactive JsonTree and not raw JSON string
    // The first object `{` should have a Chevron Down icon next to it if it's the JsonTree
    const expandToggle = page.locator('svg.lucide-chevron-down').first();
    await expect(expandToggle).toBeVisible();
    await expandToggle.click();

    // After clicking, the Chevron Right should be visible because it's collapsed
    await expect(page.locator('svg.lucide-chevron-right').first()).toBeVisible();
  });
});
