import { test, expect } from '@playwright/test';

// Use direct post to create user, ignoring seedGlobalState to avoid module resolution issue with @bufbuild
test.describe('Tools Management', () => {
  test.beforeEach(async ({ page, request }) => {
    // Login flow avoiding test-data.ts dependencies
    const BASE_URL = process.env.TEST_API_URL || 'http://localhost:50050';
    const HEADERS = {
        'Content-Type': 'application/json',
        'Authorization': 'Basic ZTJlLWFkbWluLWNvcmU6cGFzc3dvcmQ=' // e2e-admin-core:password
    };

    // Seed a specific user if not exist
    const user = {
        id: "e2e-admin-core",
        authentication: {
            basic_auth: {
                username: "e2e-admin-core",
                password_hash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a" // password
            }
        },
        roles: ["admin"],
        profile_ids: ["dev", "prod"]
    };

    try {
        await request.post('/api/v1/users', { data: user, headers: HEADERS });
    } catch(e) {}

    await page.goto('/login');
    await page.waitForSelector('form');
    await page.getByPlaceholder('e.g., admin').fill('e2e-admin-core');
    await page.getByPlaceholder('Your password').fill('password');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await page.waitForURL('**/dashboard');

    // Seed a service with a tool
    await request.post('/api/v1/services', {
      data: {
        name: 'test-service-for-tools',
        http_service: {
          address: 'http://localhost:8080',
          tools: [
            {
              name: 'test-tool-to-disable',
              description: 'A tool to test disabling',
              call_id: 'test-call'
            }
          ],
          calls: {
            'test-call': {
              id: 'test-call',
              endpoint_path: '/test',
              method: 'HTTP_METHOD_GET'
            }
          }
        },
        disable: false
      },
      headers: HEADERS
    });
  });

  test.afterEach(async ({ request }) => {
    // Cleanup
    const HEADERS = {
        'Content-Type': 'application/json',
        'Authorization': 'Basic ZTJlLWFkbWluLWNvcmU6cGFzc3dvcmQ=' // e2e-admin-core:password
    };
    await request.delete('/api/v1/services/test-service-for-tools', { headers: HEADERS });
  });

  test('should allow disabling and enabling a tool', async ({ page, request }) => {
    await page.goto('/tools');

    // Wait for the tool to appear
    await expect(page.getByText('test-tool-to-disable')).toBeVisible();

    // The tool should be Enabled by default (Switch checked = Enabled)
    const row = page.locator('tr').filter({ hasText: 'test-tool-to-disable' });
    await expect(row.getByRole('switch')).toBeChecked();
    await expect(row.getByText('Enabled')).toBeVisible();

    // Disable the tool
    await row.getByRole('switch').click();

    // Verify UI updates
    await expect(row.getByRole('switch')).not.toBeChecked();
    await expect(row.getByText('Disabled')).toBeVisible();

    // Verify backend state changed by fetching the service
    const HEADERS = {
        'Content-Type': 'application/json',
        'Authorization': 'Basic ZTJlLWFkbWluLWNvcmU6cGFzc3dvcmQ=' // e2e-admin-core:password
    };
    const response = await request.get('/api/v1/services/test-service-for-tools', { headers: HEADERS });
    expect(response.ok()).toBeTruthy();
    const serviceConfig = await response.json();

    // Check if tools map has the disabled flag
    expect(serviceConfig.httpService.tools).toBeDefined();
    expect(serviceConfig.httpService.tools.some((t: any) => t.name === 'test-tool-to-disable' && t.disable === true)).toBeTruthy();
  });
});
