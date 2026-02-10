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
    const stackName = `test-stack-${Date.now()}`;
    const yamlContent = `name: ${stackName}
description: "A test stack"
services:
  - name: test-service
    commandLineService:
      command: echo "hello"
`;

    // Use evaluate to set value in Monaco directly to avoid keyboard flakiness
    await page.evaluate((content) => {
        // @ts-ignore
        if (window.monaco && window.monaco.editor) {
            // @ts-ignore
            const models = window.monaco.editor.getModels();
            if (models.length > 0) {
                models[0].setValue(content);
            }
        }
    }, yamlContent);

    // Wait a bit for React state to sync
    await page.waitForTimeout(500);

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
