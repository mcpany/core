/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('HTTP Tool Editor', () => {
  const serviceName = 'e2e-http-tool-test';

  test.beforeEach(async ({ request }) => {
    // Clean up if exists
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
  });

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`).catch(() => {});
  });

  test('should create an HTTP service with tools', async ({ page, request }) => {
    test.slow(); // Increase timeout

    await page.goto('/upstream-services');
    await page.getByRole('button', { name: 'Add Service' }).click();

    // Select Custom Service template
    // Wait for templates to load
    await expect(page.getByText('Custom Service')).toBeVisible({ timeout: 30000 });
    await page.getByText('Custom Service').click();

    // Fill Basic Info
    await page.getByLabel('Service Name').fill(serviceName);

    // Go to Connection Tab
    await page.getByRole('tab', { name: 'Connection' }).click();
    await page.getByLabel('Base URL').fill('https://api.example.com');

    // Go to Tools Tab
    await page.getByRole('tab', { name: 'Tools' }).click();

    // Add Tool
    await page.getByRole('button', { name: 'Add Tool' }).click();

    // Tool Editor
    await page.getByLabel('Tool Name').fill('get_user');
    await page.getByLabel('Description').fill('Get user by ID');

    // Path
    await page.getByLabel('Endpoint Path').fill('/users/{id}');

    // Add Parameter
    await page.getByRole('button', { name: 'Add Parameter' }).click();

    // Fill Parameter
    await page.locator('input[placeholder="param_name"]').fill('id');
    await page.locator('input[placeholder="Description"]').fill('User ID');

    // Save Tool
    await page.getByRole('button', { name: 'Save Tool' }).click();

    // Verify tool in list
    await expect(page.getByText('get_user')).toBeVisible();
    await expect(page.getByText('/users/{id}')).toBeVisible();

    // Save Service
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Verify Persistence
    // Wait for toast
    await expect(page.getByText('Service Created').first()).toBeVisible({ timeout: 10000 });

    // API Check
    const response = await request.get(`/api/v1/services/${serviceName}`);
    expect(response.ok()).toBeTruthy();
    const service = await response.json();

    // Note: The API returns snake_case
    const httpService = service.service?.http_service || service.http_service;

    expect(httpService).toBeDefined();
    expect(httpService.tools).toHaveLength(1);
    expect(httpService.tools[0].name).toBe('get_user');

    // Verify Call Definition
    const callId = httpService.tools[0].call_id;
    expect(callId).toBeDefined();
    expect(httpService.calls[callId]).toBeDefined();
    expect(httpService.calls[callId].endpoint_path).toBe('/users/{id}');
    expect(httpService.calls[callId].parameters).toHaveLength(1);
    expect(httpService.calls[callId].parameters[0].schema.name).toBe('id');
  });
});
