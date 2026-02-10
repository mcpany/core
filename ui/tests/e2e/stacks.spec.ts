/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test('Stacks Management Flow', async ({ page, request }) => {
  // 1. Seed a stack directly via API to verify backend works and seed data
  const stackName = 'seeded-stack';
  const yaml = `name: ${stackName}
description: Seeded stack
services:
  - name: dummy-web
    http_service:
      address: http://example.com`;

  console.log('Seeding stack via API...');
  const res = await request.post(`/api/v1/stacks/${stackName}/config`, {
      headers: { 'Content-Type': 'text/plain', 'X-API-Key': 'test-token' },
      data: yaml
  });

  if (!res.ok()) {
      console.log('Seed failed:', await res.text());
  }
  expect(res.ok()).toBeTruthy();
  console.log('Seed successful.');

  // 2. Navigate to Stacks page
  await page.goto('/stacks');
  await expect(page.getByRole('heading', { name: 'Stacks' })).toBeVisible();

  // 3. Verify Seeded Stack is visible
  // Use first() to avoid strict mode violation if error toast appears with same text
  await expect(page.getByText(stackName).first()).toBeVisible();

  // 4. Click on the stack to edit
  await page.getByText(stackName).first().click();

  // 5. Verify Editor loads
  await expect(page).toHaveURL(new RegExp(`/stacks/${stackName}`));
  await expect(page.getByRole('heading', { name: `Edit Stack: ${stackName}` })).toBeVisible();

  // 6. Delete Stack via UI
  await page.goto('/stacks');
  const stackCard = page.locator('.group', { hasText: stackName }).first();
  await stackCard.hover();

  page.on('dialog', dialog => dialog.accept());

  // Click the delete button (trash icon)
  await stackCard.locator('button').filter({ has: page.locator('svg') }).click();

  // Verify it's gone
  await expect(page.locator('a.group', { hasText: stackName })).not.toBeVisible();
});
