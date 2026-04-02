import { test, expect } from '@playwright/test';

test.describe('HITL Approvals Dashboard', () => {
  test.beforeEach(async ({ request, baseURL }) => {
    // In CI, PLAYWRIGHT_BASE_URL (or baseURL) is the frontend server, and it proxies /api/v1 to the backend.
    // Wait, the API is actually mapped directly to the root in app/api.go for most things, but it's
    // accessed via /api/v1/ prefix through Vite proxy or the backend router.
    // Let's use the UI's base URL and hit the proxied API endpoint.
    // We need to pass the API Key to the backend
    await request.post(`${baseURL}/api/v1/mock/seed-hitl`, {
      data: {
        execution_id: 'test-execution-id-123',
        tool_name: 'production.drop_database',
        require_mfa: true,
      },
      headers: {
        'X-API-Key': process.env.MCPANY_API_KEY || 'test-token',
      }
    });

    await request.post(`${baseURL}/api/v1/mock/seed-hitl`, {
      data: {
        execution_id: 'test-execution-id-456',
        tool_name: 'production.restart_server',
        require_mfa: false,
      },
      headers: {
        'X-API-Key': process.env.MCPANY_API_KEY || 'test-token',
      }
    });
  });

  test('displays seeded approvals and allows approval process with MFA', async ({ page }) => {
    // Navigate to the HITL dashboard
    await page.goto('/hitl');

    // Wait for the approvals to load
    await expect(page.getByText('production.drop_database')).toBeVisible();
    await expect(page.getByText('production.restart_server')).toBeVisible();

    // Verify intent text
    await expect(page.getByText('Pending verification for sensitive tool').first()).toBeVisible();

    // Verify MFA requirement badge is visible
    await expect(page.getByText('MFA validation required for approval').first()).toBeVisible();

    // Click Approve on the item without MFA requirement
    const restartCard = page.locator('.shadow-lg').filter({ hasText: 'production.restart_server' });
    await restartCard.getByRole('button', { name: 'Approve' }).click();

    // Verify toast appears
    await expect(page.getByText('Action Approved')).toBeVisible();

    // Verify the item is removed from the dashboard
    await expect(page.getByText('production.restart_server')).not.toBeVisible();

    // Click Approve on the item with MFA requirement
    const dropCard = page.locator('.shadow-lg').filter({ hasText: 'production.drop_database' });
    await dropCard.getByRole('button', { name: 'Approve' }).click();

    // Verify the MFA dialog appears
    await expect(page.getByRole('dialog', { name: /Multi-Factor Authentication/i })).toBeVisible();

    // Enter MFA code
    await page.getByPlaceholder('Enter 6-digit code').fill('123456');

    // Submit MFA code
    await page.getByRole('button', { name: /Verify & Approve/i }).click();

    // Verify toast appears
    await expect(page.getByText('Action Approved')).toBeVisible();

    // Verify the item is removed from the dashboard
    await expect(page.getByText('production.drop_database')).not.toBeVisible();

    // Verify empty state is shown
    await expect(page.getByText('No pending approvals')).toBeVisible();
  });
});
