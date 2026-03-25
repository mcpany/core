import { seedGlobalState } from "./test-data";
/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";

test.describe("OAuth Flow Integration", () => {
  const credentialID = "cred-oauth-1";
  let callbackCalled = false;
  const credentials: any[] = [
    {
      id: credentialID,
      name: "GitHub OAuth",
      authentication: {
        oauth2: {
          clientId: { value: { plainText: "client-id" } },
          authorizationUrl: "http://127.0.0.1:38817/auth",
          tokenUrl: "http://127.0.0.1:38817/token",
          scopes: "read:user",
        },
      },
      token: null,
    },
  ];

  test.beforeEach(async ({ page, request }) => {
    await seedGlobalState(request);

    // Increase viewport height for long forms/lists
    await page.setViewportSize({ width: 1280, height: 1000 });

    callbackCalled = false;
    // Reset credentials for each test if multiple tests existed
    credentials[0].token = null;

    page.on("console", (msg) => console.log("BROWSER LOG:", msg.text()));

    // Mock service create

    // Mock templates list for marketplace

    // Mock template create/save
  });

  test("should complete the OAuth flow via Auth Wizard", async ({ page }) => {
    await page.goto("/marketplace");
    await page
      .getByRole("button", { name: "Create Config", exact: true })
      .click({ force: true });

    // Step 1: Type & Template in Wizard
    // The wizard shows "Manual / Custom" in the template selection card by default
    await page
      .getByPlaceholder("e.g. My Postgres DB")
      .fill("OAuth Test Service");
    await page
      .getByRole("button", { name: "Next", exact: true })
      .click({ force: true });

    // Step 2: Parameters (Skip or click Next)
    await page
      .getByRole("button", { name: "Next", exact: true })
      .click({ force: true });

    // Step 3: Webhooks (Skip or click Next)
    await page
      .getByRole("button", { name: "Next", exact: true })
      .click({ force: true });

    // Step 4: Authentication
    await expect(page.getByText("Select Credential for Testing")).toBeVisible({
      timeout: 15 * 1000,
    });

    await page.waitForTimeout(1000); // Wait for animations/mounting
    await expect(page.getByRole("combobox")).toBeVisible({
      timeout: 15 * 1000,
    });
    await page.getByRole("combobox").click({ force: true });

    // Ensure the option is actually in the list (or search for it if filtered)
    const githubOption = page.getByRole("option", { name: "GitHub OAuth" });
    await expect(githubOption).toBeVisible({ timeout: 10000 });
    await githubOption.click({ force: true });

    const connectButton = page.getByRole("button", {
      name: "Connect with OAuth",
    });
    await expect(connectButton).toBeVisible({ timeout: 10000 });
    await connectButton.click({ force: true });

    // Success check in callback page
    await expect(page.getByText("Authentication Successful")).toBeVisible({
      timeout: 30 * 1000,
    });
    expect(callbackCalled).toBeTruthy();

    await page.getByRole("button", { name: "Continue" }).click({ force: true });

    // BACK IN WIZARD
    // Now it should show Account Connected because we updated the mock
    await expect(page.getByText("Account Connected")).toBeVisible({
      timeout: 20 * 1000,
    });

    await page
      .getByRole("button", { name: "Next", exact: true })
      .click({ force: true }); // Go to Review
    await page.getByRole("button", { name: /Finish/ }).click({ force: true });

    await expect(page.getByRole("dialog")).toBeHidden({ timeout: 10 * 1000 });
  });
});
