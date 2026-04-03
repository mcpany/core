/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";

test.describe("Upstream Service Detail Page", () => {
  const serviceName = "e2e-detail-test-service";

  test.beforeAll(async ({ request }) => {
    // Seed the database with a test service
    // Note: We must use /api/v1/services because that's what the middleware proxies
    // and what the backend exposes (mounted at /api/v1/).
    const response = await request.post("/api/v1/services", {
      data: {
        name: serviceName,
        http_service: {
          address: "http://example.com",
          tools: [
            {
              name: "example_tool",
              description: "An example tool",
              input_schema: {
                type: "object",
                properties: {
                  test_param: {
                    type: "string",
                    description: "A test parameter",
                  },
                },
              },
            },
          ],
          prompts: [
            {
              name: "example_prompt",
              description: "An example prompt",
              arguments: [
                {
                  name: "arg1",
                  description: "First argument",
                  required: true,
                },
              ],
            },
          ],
          resources: [
            {
              name: "example_resource",
              description: "An example resource",
              uri_template: "file:///test/{name}.txt",
              mime_type: "text/plain",
            },
          ],
        },
        priority: 10,
      },
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test("should display ServiceEditor and save changes", async ({
    page,
    request,
  }) => {
    // 1. Navigate to the detail page
    await page.goto(`/upstream-services/${serviceName}`);

    // 2. Verify Page Title
    await expect(page.getByRole("heading", { level: 1 })).toContainText(
      serviceName,
    );

    // 3. Navigate to Settings tab where ServiceEditor is located
    await page.getByRole("tab", { name: "Settings" }).click();

    // 4. Verify ServiceEditor tabs are present (Evidence that ServiceEditor is used)
    // The ServiceEditor has: General, Connection, Authentication, Policies, Advanced
    await expect(page.getByRole("tab", { name: "Connection" })).toBeVisible();
    await expect(page.getByRole("tab", { name: "Policies" })).toBeVisible();

    // 5. Modify a field
    // Go to General tab (default) and change Priority
    // Note: ServiceEditor defaults to "general" tab.
    const priorityInput = page.getByLabel("Priority");
    await expect(priorityInput).toBeVisible();
    await expect(priorityInput).toHaveValue("10");

    await priorityInput.fill("5");

    // 5. Save Changes
    const saveButton = page.getByRole("button", { name: "Save Changes" });
    await saveButton.click();

    // 6. Verify Toast/Feedback
    // Use first() to avoid strict mode violation if multiple elements match (e.g. title and aria-live region)
    await expect(page.getByText("Service Updated").first()).toBeVisible();
    await expect(
      page.getByText("Configuration saved successfully").first(),
    ).toBeVisible();

    // 7. Verify Persistence via API
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const service = await response.json();
    expect(service.priority).toBe(5);
  });
});

test("expands tool row to view schema in DefinitionsTable", async ({
  page,
}) => {
  await page.goto(`/service/e2e-detail-test-service`);

  // Wait for the page to load
  await expect(
    page.getByRole("heading", { name: serviceName, exact: false }),
  ).toBeVisible({ timeout: 10000 });

  // Find the example_tool row
  const toolRow = page.locator("tr").filter({ hasText: "example_tool" });
  await expect(toolRow).toBeVisible();

  // Click the expand button (chevron)
  const expandButton = toolRow.locator("button");
  await expandButton.click();

  // Verify the expanded content shows the schema field "test_param"
  await expect(page.getByText("test_param")).toBeVisible();

  // Optional: test prompt and resource expansion
  const promptRow = page.locator("tr").filter({ hasText: "example_prompt" });
  await promptRow.locator("button").click();
  await expect(page.getByText("arg1")).toBeVisible();

  const resourceRow = page
    .locator("tr")
    .filter({ hasText: "example_resource" });
  await resourceRow.locator("button").click();
  await expect(page.getByText("file:///test/{name}.txt")).toBeVisible();

  test('expands tool row to view schema in DefinitionsTable', async ({ page }) => {
    await page.goto(`/service/${serviceName}`);

    // Wait for the page to load
    await expect(page.getByRole("heading", { name: serviceName, exact: false })).toBeVisible({ timeout: 10000 });

    // Find the example_tool row
    const toolRow = page.locator('tr').filter({ hasText: 'example_tool' });
    await expect(toolRow).toBeVisible();

    // Click the expand button (chevron)
    const expandButton = toolRow.locator('button');
    await expandButton.click();

    // Verify the expanded content shows the schema field "test_param"
    await expect(page.getByText('test_param')).toBeVisible();

    // Optional: test prompt and resource expansion
    const promptRow = page.locator('tr').filter({ hasText: 'example_prompt' });
    await promptRow.locator('button').click();
    await expect(page.getByText('arg1')).toBeVisible();

    const resourceRow = page.locator('tr').filter({ hasText: 'example_resource' });
    await resourceRow.locator('button').click();
    await expect(page.getByText('file:///test/{name}.txt')).toBeVisible();
  });
});
