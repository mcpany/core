/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Stacks Management', () => {
  test('should create, edit, and list a stack', async ({ page }) => {
    // 1. Navigate to Stacks page
    await page.goto('/stacks');
    await expect(page.getByRole('heading', { name: 'Stacks' })).toBeVisible();

    // 2. Click "New Stack"
    await page.getByRole('button', { name: 'New Stack' }).click();
    await expect(page).toHaveURL(/\/stacks\/new/);
    await expect(page.getByRole('heading', { name: 'New Stack' })).toBeVisible();

    // 3. Enter YAML content
    // Note: Monaco editor is tricky to type into with Playwright.
    // We try to click the editor and type.
    const editorLocator = page.locator('.monaco-editor').first();
    await expect(editorLocator).toBeVisible();
    await editorLocator.click();

    // Clear existing content (Ctrl+A, Del) - might need Command+A on Mac but Control+A usually works in browser automation context or we can force it.
    await page.keyboard.press('Control+A');
    await page.keyboard.press('Delete');

    const stackName = `test-stack-${Date.now()}`;
    const yamlContent = `name: ${stackName}
description: "A test stack"
services:
  - name: test-service
    commandLineService:
      command: echo "hello"
`;
    await page.keyboard.insertText(yamlContent);

    // 4. Save
    await page.getByRole('button', { name: 'Save Stack' }).click();

    // 5. Verify redirection to detail page (URL should contain stack name)
    await expect(page).toHaveURL(new RegExp(`/stacks/${stackName}`));
    await expect(page.getByRole('heading', { name: stackName })).toBeVisible();

    // 6. Navigate back to list
    await page.goto('/stacks');

    // 7. Verify stack is in list
    await expect(page.getByText(stackName)).toBeVisible();
    // Description might be truncated or hidden depending on layout, but let's check
    await expect(page.getByText("A test stack")).toBeVisible();
  });
});
