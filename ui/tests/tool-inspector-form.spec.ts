import { test, expect } from '@playwright/test';

const SERVICE_ID = 'test-form-service';
const TOOL_NAME = 'test_form_tool';
// The backend namespaces the tool name with the service ID
const DISPLAY_NAME = `${SERVICE_ID}.${TOOL_NAME}`;

test.describe('Tool Inspector Dynamic Form', () => {

  test.beforeAll(async ({ request }) => {
    // Clean up if exists
    await request.delete(`/api/v1/services/${SERVICE_ID}`).catch(() => {});

    // Register service with complex tool schema
    const response = await request.post('/api/v1/services', {
      data: {
        id: SERVICE_ID,
        name: SERVICE_ID,
        version: "1.0.0",
        http_service: {
            address: "https://httpbin.org",
            tools: [
                {
                    name: TOOL_NAME,
                    description: "A tool to test the dynamic form",
                    call_id: "test_call"
                }
            ],
            calls: {
                test_call: {
                    endpoint_path: "/get",
                    method: "HTTP_METHOD_GET",
                    input_schema: {
                        type: "object",
                        properties: {
                            stringField: {
                                type: "string",
                                description: "A simple string"
                            },
                            numberField: {
                                type: "number",
                                description: "A simple number"
                            },
                            booleanField: {
                                type: "boolean",
                                description: "A switch"
                            },
                            enumField: {
                                type: "string",
                                enum: ["Option A", "Option B", "Option C"],
                                description: "A dropdown"
                            },
                            nested: {
                                type: "object",
                                properties: {
                                    childField: { type: "string" }
                                }
                            }
                        },
                        required: ["stringField"]
                    }
                }
            }
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
     await request.delete(`/api/v1/services/${SERVICE_ID}`).catch(() => {});
  });

  test('should render dynamic form and execute tool', async ({ page }) => {
    await page.goto('/tools');

    // Filter to find the tool easily
    await page.getByPlaceholder('Search tools...').fill(TOOL_NAME);
    // Wait for the tool to appear in the table
    await expect(page.getByRole('cell', { name: DISPLAY_NAME, exact: true })).toBeVisible();

    // Open Inspector
    // Locate the row containing the tool name, then find the 'Inspect' button within it
    const row = page.getByRole('row').filter({ has: page.getByRole('cell', { name: DISPLAY_NAME, exact: true }) });
    await row.getByRole('button', { name: 'Inspect' }).click();

    // Verify Inspector is open
    await expect(page.getByRole('dialog')).toBeVisible();
    // The dialog title might also use the display name or just the tool name depending on implementation.
    // ToolInspector title: {String(tool.name)}
    // So it should be DISPLAY_NAME.
    await expect(page.getByRole('heading', { name: DISPLAY_NAME })).toBeVisible();

    // Verify Form Fields
    await expect(page.locator('label').filter({ hasText: 'stringField' })).toBeVisible();
    await expect(page.locator('label').filter({ hasText: 'numberField' })).toBeVisible();
    await expect(page.locator('label').filter({ hasText: 'booleanField' })).toBeVisible();
    await expect(page.locator('label').filter({ hasText: 'enumField' })).toBeVisible();
    await expect(page.locator('label').filter({ hasText: 'childField' })).toBeVisible();

    // Fill Form
    // Using nth(0) if multiple inputs match, but better to target by sibling of label?
    // DynamicForm doesn't associate label `for` with input `id` perfectly (nested structure).
    // But basic inputs are children of labels div? No, siblings.
    // The Input inside DynamicForm doesn't have ID unless passed.
    // My DynamicForm implementation did NOT pass unique IDs to Inputs (except type=number placeholder).
    // But I can target by placeholder "Enter text..." or "Enter number..."

    // Wait, let's look at DynamicForm implementation again.
    // It renders Label then Input.

    // Fill String Field
    // We can use layout selectors.
    await page.locator('div').filter({ hasText: /^stringField/ }).getByPlaceholder('Enter text...').fill('Hello World');

    // Fill Number Field
    await page.locator('div').filter({ hasText: /^numberField/ }).getByPlaceholder('Enter number...').fill('42');

    // Fill Enum (Select)
    await page.locator('div').filter({ hasText: /^enumField/ }).getByRole('combobox').click();
    await page.getByLabel('Option B').click();

    // Fill Nested
    await page.locator('div').filter({ hasText: /^childField/ }).getByPlaceholder('Enter text...').fill('Nested Value');

    // Toggle Boolean (Switch)
    await page.locator('div').filter({ hasText: /^booleanField/ }).getByRole('switch').click();

    // Switch to JSON view to verify
    await page.getByRole('tablist').filter({ hasText: 'Form' }).getByRole('tab', { name: 'JSON' }).click();

    // Check JSON content
    const jsonContent = await page.locator('textarea#args').inputValue();
    const parsed = JSON.parse(jsonContent);

    expect(parsed.stringField).toBe('Hello World');
    expect(parsed.numberField).toBe(42);
    expect(parsed.enumField).toBe('Option B');
    expect(parsed.booleanField).toBe(true);
    expect(parsed.nested.childField).toBe('Nested Value');

    // Switch back to Form
    await page.getByRole('tab', { name: 'Form', exact: true }).click();

    // Execute and spy on network request to verify payload
    const executeRequestPromise = page.waitForRequest(req => req.url().includes('/api/v1/execute') && req.method() === 'POST');
    await page.getByRole('button', { name: 'Execute' }).click();
    const request = await executeRequestPromise;
    const postData = request.postDataJSON();

    // Verify arguments sent to backend match form data
    expect(postData.arguments).toEqual({
        stringField: 'Hello World',
        numberField: 42,
        booleanField: true,
        enumField: 'Option B',
        nested: { childField: 'Nested Value' }
    });

    // Verify Result shows up (even if error)
    await expect(page.getByText('Result')).toBeVisible();
  });
});
