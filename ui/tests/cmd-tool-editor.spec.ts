/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Command Line Tool Editor', () => {
  const serviceName = 'e2e-cmd-tool-test';

  test.beforeAll(async ({ request }) => {
    // Seed CMD service
    // Ensure cleanup first just in case
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});

    const response = await request.post('/api/v1/services', {
      data: {
        name: serviceName,
        command_line_service: {
            command: "echo"
        }
      }
    });
    expect(response.ok()).toBeTruthy();
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`);
  });

  test('should define tools for Command Line service', async ({ page, request }) => {
    // Navigate to upstream services
    await page.goto('/upstream-services');

    // Wait for list to load
    await expect(page.getByText(serviceName)).toBeVisible();

    // Find row, click edit.
    const row = page.getByRole('row', { name: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Wait for Sheet to open.
    await expect(page.getByRole('dialog', { name: 'Edit Service' })).toBeVisible();

    // Click "Tools" tab.
    await page.getByRole('tab', { name: 'Tools' }).click();

    // Click "Add Tool"
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // Sheet for Tool Editor should open
    await expect(page.getByText('Edit new_tool')).toBeVisible();

    // Fill details
    await page.getByLabel('Tool Name').fill('say_hello');
    await page.getByLabel('Description').fill('Say hello to someone');

    // Add Argument
    await page.getByRole('button', { name: 'Add Argument' }).click();

    // We need to target the argument input. Since there's drag-and-drop, it might be tricky.
    // The placeholder is "Argument 1"
    await page.getByPlaceholder('Argument 1').fill('Hello {{ msg }}');

    // Scan Args
    await page.getByRole('button', { name: 'Scan Args' }).click();

    // Check if parameter was added
    // Use locator with value attribute as fallback if getByDisplayValue is somehow problematic
    await expect(page.locator('input[value="msg"]')).toBeVisible();
    await expect(page.locator('input[value="Auto-generated from argument"]')).toBeVisible();

    // Close Tool Editor Sheet (press Escape)
    await page.keyboard.press('Escape');
    await expect(page.getByRole('heading', { name: 'Edit say_hello' })).not.toBeVisible();

    // Verify parent sheet (Service Editor) is still open
    await expect(page.getByRole('dialog', { name: 'Edit Service' })).toBeVisible();

    // Verify tool is listed in Manager within the parent sheet
    // It should show tool name and preview of args
    await expect(page.getByRole('dialog', { name: 'Edit Service' }).getByText('say_hello', { exact: true })).toBeVisible();
    // Preview might be "echo Hello {{ msg }}"
    await expect(page.getByRole('dialog', { name: 'Edit Service' }).getByText('echo Hello {{ msg }}')).toBeVisible();

    // Save Service
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Verify Toast
    await expect(page.getByText('Service Updated', { exact: true })).toBeVisible();

    // Verify Persistence via API
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const service = await response.json();
    const cmdService = service.command_line_service || service.commandLineService;

    expect(cmdService.tools).toHaveLength(1);
    expect(cmdService.tools[0].name).toBe('say_hello');

    // Verify Call Definition
    const callId = cmdService.tools[0].callId || cmdService.tools[0].call_id;
    const calls = cmdService.calls; // Map

    // Check if calls is populated
    expect(calls).toBeDefined();
    const call = calls[callId];
    expect(call).toBeDefined();

    expect(call.args).toContain('Hello {{ msg }}');
    expect(call.parameters).toHaveLength(1);
    // Check param name inside schema
    const param = call.parameters[0];
    const paramName = param.schema?.name;
    expect(paramName).toBe('msg');
  });
});
