/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";
import { seedGlobalState } from "./test-data";

test.describe("Resource Preview Modal", () => {
  test.beforeEach(async ({ request }) => {
    // We will use seedGlobalState from test-data.ts, which seeds multiple services,
    // including Echo Service (svc_echo) which has config.json.
    await seedGlobalState(request);
  });

  test("should open resource in modal from explorer", async ({ page }) => {
    await page.goto("/resources");

    // Wait for resources to load and click
    // We use config.json from svc_echo which is seeded in test-data.ts
    const resourceItem = page.locator("div.font-medium", {
      hasText: "config.json",
    });
    await expect(resourceItem).toBeVisible();

    // Click on the resource to select it
    await resourceItem.click();

    // Wait for the inline preview to render before opening the modal.
    // For JSON, we use JsonView which renders text content (keys/values).
    // config.json has text_content: '{\n  "foo": "bar"\n}'
    await expect(page.getByText("bar").first()).toBeVisible();

    // Wait for "Maximize" button and click it
    await page.click('button[title="Maximize"]');

    // Wait for modal to open and verify title
    const modalTitle = page
      .locator("div[role='dialog']")
      .getByRole("heading", { name: "config.json" });
    await expect(modalTitle).toBeVisible();

    // Verify content in modal
    const modalContentContainer = page.locator("div[role='dialog']");
    await expect(modalContentContainer.getByText("bar").first()).toBeVisible();
  });
});
