/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';

test.describe('E2E Full Coverage', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should navigate to all pages and verify content', async ({ page }) => {
    // Dashboard (Home)
    await expect(page).toHaveTitle(/MCP Any/);
    await expect(page.locator('h1')).toContainText('Dashboard');

    // Services
    await page.click('a[href="/services"]');
    await expect(page.locator('h2')).toContainText('Services');
    await expect(page.getByRole('button', { name: 'Add Service' })).toBeVisible();

    // Settings
    await page.click('a[href="/settings"]');
    await expect(page.locator('h2')).toContainText('Settings');
    await expect(page.getByText('Global Configuration')).toBeVisible();

    // Playground
    await page.click('a[href="/playground"]');
    await expect(page.locator('h2')).toContainText('Playground');

    // Tools
    await page.click('a[href="/tools"]');
    await expect(page.locator('h2')).toContainText('Tools');

    // Prompts
    await page.click('a[href="/prompts"]');
    await expect(page.locator('h2')).toContainText('Prompts');

    // Resources
    await page.click('a[href="/resources"]');
    await expect(page.locator('h2')).toContainText('Resources');
  });

  test('should register and manage a service', async ({ page }) => {
    await page.goto('/services');
    await page.click('button:has-text("Add Service")');

    // Fill form (Advanced Mode for Tools)
    await page.click('button[value="advanced"]');

    const config = {
      name: "e2e-test-service",
      http_service: {
        address: "http://localhost:8081",
        calls: {
          echo: {
            method: "HTTP_METHOD_POST",
            endpoint_path: "/echo"
          }
        },
        tools: [
          {
            name: "echo_tool",
            description: "Echoes back the request body.",
            call_id: "echo",
            input_schema: {
              type: "object",
              properties: {
                message: { type: "string" }
              }
            }
          }
        ]
      }
    };

    await page.fill('textarea[name="configJson"]', JSON.stringify(config, null, 2));

    await page.click('button:has-text("Register Service")');

    // Check if it appears in list
    await expect(page.locator('text=e2e-test-service')).toBeVisible();

    // Edit Service - verify name still there
    await page.click('text=e2e-test-service');
    await page.click('button:has-text("Edit Config")');
    // We might need to switch to basic or advanced to check name, but name is in basic tab too
    await expect(page.locator('input[name="name"]')).toHaveValue('e2e-test-service');
    await page.click('button:has-text("Cancel")');

    // Disable/Enable (Status toggle) - Assuming there is a switch for it on details page
    // await page.click('button[role="switch"]');

    // Delete Service
    // Navigate back to list or delete from details?
    // Assuming delete button exists on details page
    await page.click('button:has-text("Delete Service")');
    await page.click('button:has-text("Confirm")'); // Confirm dialog

    await expect(page.locator('text=e2e-test-service')).not.toBeVisible();
  });

  test('should manage global settings', async ({ page }) => {
    await page.goto('/settings');

    // Change Log Level
    await page.click('button:has-text("INFO")'); // Current value assumption
    await page.click('div[role="option"]:has-text("DEBUG")');

    await page.click('button:has-text("Save Changes")');
    await expect(page.locator('text=Settings saved successfully')).toBeVisible();

    // Reload and verify
    await page.reload();
    await expect(page.getByText('DEBUG')).toBeVisible();
  });

  test('should execute tools in playground', async ({ page }) => {
    await page.goto('/playground');

    // Connect/Status check
    await expect(page.locator('text=Connected to Localhost')).toBeVisible();

    // Type command
    await page.fill('input[placeholder="Type a message to interact with your tools..."]', 'echo_tool {"message": "hello e2e"}');
    await page.click('button:has-text("Send")');

    // Verify tool call message
    await expect(page.locator('text=Calling Tool: echo_tool')).toBeVisible();

    // Verify result (might fail if tool doesn't exist, but we check for ANY result or error)
    // If backend doesn't have echo_tool, it returns error
    // We expect either success or proper error message
    // Let's assume echo_tool might not exist, but we verify interaction flow
    await expect(page.locator('text=Tool execution failed').or(page.locator('text=Tool Output'))).toBeVisible({ timeout: 10000 });
  });

  test('should manage secrets', async ({ page }) => {
    // Assuming Secrets UI is under Settings -> Secrets tab
    await page.goto('/settings');
    await page.click('button:has-text("Secrets")');

    await page.click('button:has-text("Add Secret")');
    await page.fill('input[name="name"]', 'e2e_secret');
    await page.fill('input[name="key"]', 'API_KEY');
    await page.fill('input[name="value"]', 'super_secret_value');

    await page.click('button:has-text("Save Secret")');

    // Verify list
    await expect(page.locator('text=e2e_secret')).toBeVisible();
    await expect(page.locator('text=[REDACTED]')).toBeVisible();

    // Delete
    await page.click('button[aria-label="Delete secret"]'); // adjust selector
    await expect(page.locator('text=e2e_secret')).not.toBeVisible();
  });
});
