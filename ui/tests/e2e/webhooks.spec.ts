import { test, expect, request } from '@playwright/test';

test.describe('Webhooks Configuration', () => {
  test('should allow configuring alert and audit webhooks', async ({ page }) => {
    // Generate random URLs to avoid conflicts with previous runs
    const randomId = Math.random().toString(36).substring(7);
    const initialAlertUrl = `https://initial-alert.example.com/${randomId}`;
    const alertUrl = `https://alert.example.com/${randomId}`;
    const auditUrl = `https://audit.example.com/${randomId}`;

    const backendUrl = process.env.BACKEND_URL || 'http://localhost:50050';

    // Seed data via API
    const apiContext = await request.newContext({
        extraHTTPHeaders: {
            'X-API-Key': process.env.MCPANY_API_KEY || 'test-token',
        }
    });

    // Seed initial state
    await apiContext.post(`${backendUrl}/api/v1/alerts/webhook`, {
        data: { url: initialAlertUrl }
    });

    await page.goto('/webhooks');

    // 1. Configure Alert Webhook
    await expect(page.locator('h1')).toContainText('Webhooks');

    // Check seeded value
    await expect(page.locator('#alert-url')).toHaveValue(initialAlertUrl);

    // Fill New Alert URL
    await page.fill('#alert-url', alertUrl);

    // Click Save in the "System Alerts" card
    // Using accessible locators within the card context
    // Shadcn cards usually have a header title.
    const alertCard = page.locator('.rounded-xl, .border', { hasText: 'System Alerts' }).first();
    await alertCard.getByRole('button', { name: 'Save Configuration' }).click();

    // Wait for toast
    await expect(page.getByText('Alert webhook URL updated successfully')).toBeVisible();

    // 2. Configure Audit Webhook
    const auditCard = page.locator('.rounded-xl, .border', { hasText: 'Audit Log Stream' }).first();

    // Enable switch
    const switchEl = page.locator('#audit-enabled');
    const isChecked = await switchEl.isChecked();
    if (!isChecked) {
        await switchEl.click();
    }

    // Fill Audit URL
    await page.fill('#audit-url', auditUrl);

    // Click Save
    await auditCard.getByRole('button', { name: 'Save Configuration' }).click();

    // Wait for toast
    await expect(page.getByText('Audit webhook settings updated successfully')).toBeVisible();

    // 3. Verify Persistence via API
    // We verify via API to ensure the backend actually received and stored the values
    const alertRes = await apiContext.get(`${backendUrl}/api/v1/alerts/webhook`);
    const alertData = await alertRes.json();
    expect(alertData.url).toBe(alertUrl);

    const settingsRes = await apiContext.get(`${backendUrl}/api/v1/settings`);
    const settingsData = await settingsRes.json();
    expect(settingsData.audit.webhook_url).toBe(auditUrl);
    // Check storage type (4 or "STORAGE_TYPE_WEBHOOK")
    expect([4, "STORAGE_TYPE_WEBHOOK"]).toContain(settingsData.audit.storage_type);
  });
});
