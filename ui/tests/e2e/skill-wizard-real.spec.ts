/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './test-data';

test.describe('Skill Wizard (Real Data)', () => {
  test.beforeEach(async ({ request }) => {
    // Seed backend with services that have tools (e.g. Echo Service has echo_tool)
    await seedServices(request);
  });

  test.afterEach(async ({ request }) => {
    await cleanupServices(request);
  });

  test('should load tools from backend and allow selection', async ({ page }) => {
    // Navigate to Skill Creation Page
    await page.goto('/skills/create');

    // Check header
    await expect(page.getByRole('heading', { name: 'Create New Skill' })).toBeVisible();

    // Verify step 1 inputs
    await page.fill('input#name', 'test-skill');
    await page.fill('textarea#description', 'A test skill');

    // Click MultiSelect trigger (combobox)
    // The placeholder text "Select allowed tools..." should be visible
    const trigger = page.getByRole('combobox');
    await expect(trigger).toBeVisible();
    await trigger.click();

    // Verify "echo_tool" (from seedServices) is in the list
    // The Command list might be in a Popover which is attached to body.
    // We search for the option text.
    // seedServices adds "Echo Service" with tool "echo_tool"
    await expect(page.getByRole('option', { name: 'echo_tool' })).toBeVisible();

    // Select "echo_tool"
    await page.getByRole('option', { name: 'echo_tool' }).click();

    // Verify it appears as a badge in the trigger
    // The trigger button should contain the text "echo_tool"
    await expect(trigger).toHaveText(/echo_tool/);

    // Continue to next step (optional validation)
    await page.getByRole('button', { name: 'Next' }).click();
    // Should be on Step 2
    await expect(page.getByText('Step 2: Instructions')).toBeVisible();
  });
});
