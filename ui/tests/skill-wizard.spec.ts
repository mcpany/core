/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Skill Wizard', () => {
  test('should create a skill with selected tools', async ({ page }) => {
    // Mock available tools
    await page.route('**/api/v1/tools', async route => {
      const json = {
        tools: [
          { name: 'add', description: 'Add numbers' },
          { name: 'subtract', description: 'Subtract numbers' }
        ]
      };
      await route.fulfill({ json });
    });

    // Mock Skills API (List and Create)
    await page.route('**/api/v1/skills', async route => {
        if (route.request().method() === 'GET') {
             await route.fulfill({ json: { skills: [{ name: 'calculator-skill-e2e', allowedTools: ['add', 'subtract'] }] } });
        } else if (route.request().method() === 'POST') {
             await route.fulfill({ status: 200, json: { skill: { name: 'calculator-skill-e2e' } } });
        } else {
             await route.continue();
        }
    });

    await page.goto('/skills/create');

    // Fill Basic Info
    await page.fill('input#name', 'calculator-skill-e2e');
    await page.fill('textarea#description', 'A skill that calculates things.');

    // Select Tools
    await page.click('button[role="combobox"]');

    // Check tools visible
    await expect(page.getByRole('option', { name: 'add' })).toBeVisible();

    // Select tools
    await page.getByRole('option', { name: 'add' }).click();
    await page.getByRole('option', { name: 'subtract' }).click();

    // Close popover
    await page.keyboard.press('Escape');

    // Verify badges are inside the combobox button
    const combobox = page.locator('button[role="combobox"]');
    await expect(combobox).toContainText('add');
    await expect(combobox).toContainText('subtract');

    // Next Steps
    await page.click('button:has-text("Next")');
    await page.fill('textarea', 'Use the add tool.');
    await page.click('button:has-text("Next")');

    await page.click('button:has-text("Create Skill")');

    // Verify Redirect
    await expect(page).toHaveURL(/\/skills$/);

    // Verify List
    await expect(page.locator('text=calculator-skill-e2e')).toBeVisible();
    await expect(page.locator('text=2 Tools')).toBeVisible();
  });
});
