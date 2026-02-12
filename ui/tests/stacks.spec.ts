/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Stacks Management', () => {
  test('should create, edit, and delete a stack', async ({ page }) => {
    page.on('console', msg => console.log(`BROWSER LOG: ${msg.text()}`));

    // 1. Navigate to Stacks page
    await page.goto('/stacks');
    await expect(page.locator('h1')).toContainText('Stacks');

    // 2. Click "Create Stack"
    await page.getByRole('link', { name: 'Create Stack' }).first().click();
    await page.waitForURL('**/stacks/new');

    // 3. Enter YAML content
    await page.waitForSelector('.monaco-editor');

    const stackName = `e2e-stack-${Date.now()}`;
    const yamlContent = `name: "${stackName}"
version: 1.0.0
services:
- name: "${stackName}-s1"
  command_line_service: { command: ls }`;

    await page.click('.monaco-editor');
    await page.keyboard.press('Control+A');
    await page.keyboard.press('Delete');
    await page.keyboard.insertText(yamlContent);

    // 4. Save
    await page.click('text=Save & Deploy');

    // 5. Verify redirection
    await expect(page).toHaveURL(new RegExp(`/stacks/${stackName}`));
    await expect(page.locator('h1')).toContainText(`Edit Stack: ${stackName}`);

    // 6. Navigate back to list
    await page.goto('/stacks');
    // Ensure the stack is present
    await expect(page.locator('.grid > a', { hasText: stackName })).toBeVisible();

    // 7. Delete stack
    page.on('dialog', dialog => dialog.accept());

    const cardLink = page.locator('.grid > a', { hasText: stackName });
    // We need to click the delete button which is inside the card.
    // The card is the link. The delete button stops propagation?
    // In page.tsx: onClick={(e) => handleDelete(e, stack.name)} and e.preventDefault().
    const deleteBtn = cardLink.getByRole('button');
    await deleteBtn.click();

    // 8. Verify removal
    // Check that the card is gone. We use the specific card locator.
    await expect(page.locator('.grid > a', { hasText: stackName })).not.toBeVisible();
  });
});
