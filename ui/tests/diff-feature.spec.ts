/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";

test.describe("Playground Tool Output Diffing", () => {
  const serviceName = "diff-feature-test-service";

  test.beforeAll(async ({ request }) => {
    // Clean up
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    // Seed service
    const response = await request.post("/api/v1/services", {
      data: {
        name: serviceName,
        command_line_service: {
          command: "echo",
          tools: [
            {
              name: "diff_test_tool",
              call_id: "call1",
              description: "Test diffing",
              inputSchema: {
                type: "object",
                properties: { arg: { type: "string" } },
              },
            },
          ],
          calls: {
            call1: {
              args: [JSON.stringify({ value: "{{.arg}}" })],
            },
          },
        },
      },
    });
    if (!response.ok()) {
      console.log(await response.text());
    }
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
  });

  test("should allow comparing tool outputs when they differ", async ({
    page,
  }) => {
    await page.goto("/playground");

    // 1. Run the tool first time
    await page.fill(
      'input[placeholder="Enter command or select a tool..."]',
      `${serviceName}.diff_test_tool {"arg":"Version 1"}`,
    );
    await page.keyboard.press("Enter");

    // Wait for first result
    await expect(page.getByText('"Version 1"')).toBeVisible();

    // 2. Run the tool second time (same args)
    // The input clears after send, so we type again.
    await page.fill(
      'input[placeholder="Enter command or select a tool..."]',
      `${serviceName}.diff_test_tool {"arg":"Version 2"}`,
    );
    await page.keyboard.press("Enter");

    // Wait for second result
    await expect(page.getByText('"Version 2"')).toBeVisible();

    // 3. Check for "Show Changes" button
    // It SHOULD be visible now.
    const showDiffBtn = page.getByRole("button", { name: "Show Changes" });
    await expect(showDiffBtn).toBeVisible();

    // 4. Click the button
    await showDiffBtn.click();

    // 5. Verify Dialog opens and Diff Editor is present
    await expect(page.getByRole("dialog")).toBeVisible();
    await expect(page.getByText("Output Difference")).toBeVisible();

    // Check for Monaco Diff Editor. It usually has a class 'monaco-diff-editor'.
    // Or we can check for the content text being present twice (original and modified).
    // Monaco renders text in lines.
    await expect(page.locator(".monaco-diff-editor")).toBeVisible();
  });
});
