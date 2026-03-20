import { test, expect } from '@playwright/test';

test.describe('Tools Management', () => {
  const serviceName = 'test-tool-mgmt-svc';
  const toolName = 'echo';

  test.beforeEach(async ({ request }) => {
    // 1. Delete service if it exists to ensure a clean state
    await request.delete(`/api/v1/services/${serviceName}`, {
        headers: { 'Authorization': 'Bearer test-key' }
    });

    // 2. Seed the database with a command-line service that has an explicit tool
    const config = {
      name: serviceName,
      command_line_service: {
        tools: [
          {
            name: toolName,
            disable: false,
            description: "An echo tool"
          }
        ]
      }
    };

    const response = await request.post('/api/v1/services', {
      data: config,
      headers: {
        'Authorization': 'Bearer test-key',
        'Content-Type': 'application/json'
      }
    });

    expect(response.ok()).toBeTruthy();

    // Give backend a moment to reload config
    await new Promise(r => setTimeout(r, 1000));
  });

  test.afterEach(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceName}`, {
        headers: { 'Authorization': 'Bearer test-key' }
    });
  });

  test('can toggle tool disable state and persist it', async ({ page, request }) => {
    // Navigate to Tools page
    await page.goto('/tools');

    // Wait for the tool to be present in the table
    const toolRow = page.locator(`tr:has-text("${toolName}")`);
    await expect(toolRow).toBeVisible({ timeout: 10000 });

    // Verify it is enabled initially
    const toggleSwitch = toolRow.locator('button[role="switch"]');
    await expect(toggleSwitch).toBeVisible();
    await expect(toggleSwitch).toHaveAttribute('aria-checked', 'true');

    // Click toggle to disable
    await toggleSwitch.click();

    // Wait a moment for network roundtrip to finish saving
    await page.waitForTimeout(500);

    // Verify UI updates optimistic state
    await expect(toggleSwitch).toHaveAttribute('aria-checked', 'false');

    // Verify the backend API reflects the change (The database was seeded and actually saved)
    const toolsRes = await request.get('/api/v1/tools', {
        headers: { 'Authorization': 'Bearer test-key' }
    });
    expect(toolsRes.ok()).toBeTruthy();
    const tools = await toolsRes.json();
    const myTool = tools.find((t: any) => t.name === toolName && t.serviceId === serviceName);

    expect(myTool).toBeDefined();
    expect(myTool.disable).toBe(true);

    // Reload the page to verify it persisted across sessions
    await page.reload();
    const reloadedRow = page.locator(`tr:has-text("${toolName}")`);
    await expect(reloadedRow).toBeVisible();
    const reloadedToggle = reloadedRow.locator('button[role="switch"]');
    await expect(reloadedToggle).toHaveAttribute('aria-checked', 'false');
  });
});
