import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser, seedProfiles, cleanupProfiles, seedServices, cleanupServices } from '../e2e/test-data';

test.describe('RichResultViewer E2E', () => {
  const serviceId = 'e2e-rich-result-service';

  test.beforeEach(async ({ request, page }) => {
    await cleanupUser(request, "test-api-user").catch(() => { });
    await seedProfiles(request);
    await seedUser(request, "e2e-admin-users");
    // Seed a mock service with a tool that returns a single object
    const serviceConfig = {
      id: serviceId,
      name: serviceId,
      version: "1.0",
      command_line_service: {
        command: "node",
        args: ["-e", "console.log(JSON.stringify({jsonrpc:'2.0',id:1,result:{tools:[{name:'return_object',description:'Returns a single object'}],callResult:{key1:'value1',key2:42}}}));"]
      }
    };
    await cleanupServices(request).catch(() => {});
    await seedServices(request);

    // Seed our specific test service
    await request.post('/api/v1/services', {
      data: serviceConfig,
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'test-token'
      }
    });

    await page.goto('/login');
    await page.fill('input[name="username"]', 'e2e-admin-users');
    await page.fill('input[name="password"]', 'password');
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
    await expect(page).toHaveURL('/', { timeout: 15000 });
  });

  test.afterEach(async ({ request }) => {
    // Delete our specific test service
    await request.delete(`/api/v1/services/${serviceId}`, {
      headers: {
        'X-API-Key': 'test-token'
      }
    }).catch(() => {});
    await cleanupUser(request, "e2e-admin-users").catch(() => { });
    await cleanupUser(request, "test-api-user").catch(() => { });
    await cleanupProfiles(request);
    await cleanupServices(request).catch(() => {});
  });

  test('should render a single object result as a Key-Value table', async ({ page }) => {
    await page.goto('/playground');

    // Make sure we're on the playground and it loaded
    await expect(page.getByRole('heading', { name: 'Playground' })).toBeVisible({ timeout: 15000 });

    // Select our specific tool
    const toolRow = page.getByText(serviceId, { exact: false }).locator('..').locator('button', { hasText: 'Use' });
    await toolRow.first().click();

    // Execute command in Tool Runner
    await page.getByRole('button', { name: 'Execute', exact: true }).click();

    // Wait for the result
    // Since we patched RichResultViewer to show a "Table" tab for objects
    // we just need to verify the Table tab is visible
    const tableTab = page.getByRole('tab', { name: /Table/i }).first();
    await expect(tableTab).toBeVisible({ timeout: 15000 });

    // And verify the Key / Value headers are present
    await expect(page.getByRole('columnheader', { name: 'Key' })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Value' })).toBeVisible();
  });
});
