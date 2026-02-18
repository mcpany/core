import { test, expect } from '@playwright/test';

test.describe('Wizard E2E', () => {
  // Use a unique name to avoid conflicts if backend state persists
  const serviceName = `postgres-wizard-${Date.now()}`;

  test('creates a PostgreSQL service using the wizard', async ({ page, request }) => {
    // 1. Navigate to Marketplace
    await page.goto('/marketplace');

    // 2. Open Wizard
    await page.click('button:has-text("Create Config")');
    await expect(page.getByText('Create Upstream Service Config')).toBeVisible();

    // 3. Select Template
    // The Select Trigger has id="service-template"
    await page.click('button#service-template');
    await page.click('div[role="option"]:has-text("PostgreSQL Database")');

    // 4. Fill Service Name
    await page.fill('input#service-name', serviceName);

    // 5. Next Step (Parameters)
    await page.click('button:has-text("Next")');

    // 6. Verify Schema Form
    // Should see "Service Configuration" and "Connection URL"
    await expect(page.getByText('Service Configuration')).toBeVisible();
    await expect(page.getByLabel('Connection URL')).toBeVisible();
    await expect(page.getByLabel('Connection URL')).toHaveValue('postgresql://user:password@localhost:5432/dbname'); // Default value

    // 7. Update Parameter
    await page.fill('input[type="password"]', 'postgresql://user:pass@db:5432/testdb');

    // 8. Next Step (Webhooks - Skip)
    await page.click('button:has-text("Next")');

    // 9. Next Step (Auth - Skip)
    await page.click('button:has-text("Next")');

    // 10. Review & Finish
    await expect(page.getByText('Review & Finish')).toBeVisible();
    // Verify JSON preview contains the new URL (in plain text or masked?)
    // The review step usually shows JSON.
    // Let's just click Create.
    await page.click('button:has-text("Create Service")');

    // 11. Verify Toast
    await expect(page.getByText('Config Saved')).toBeVisible();

    // 12. Verify Backend State
    // We use the 'request' fixture to query the API directly
    const response = await request.get(`/api/v1/templates`);
    expect(response.ok()).toBeTruthy();
    const templates = await response.json();
    // Note: The wizard saves to "Backend Templates" (apiClient.saveTemplate).
    // It does NOT instantiate the service immediately (apiClient.registerService).
    // The code in MarketplacePage `handleWizardComplete` calls `apiClient.saveTemplate`.
    // Wait, let's check `marketplace/page.tsx` again.

    // In `ui/src/app/marketplace/page.tsx`:
    // const handleWizardComplete = async (config: UpstreamServiceConfig) => {
    //       await apiClient.saveTemplate(config);
    //       toast({ title: "Config Saved", description: `${config.name} saved to Backend Templates.` });
    //       setIsWizardOpen(false);
    //       loadData();
    // };

    // So it saves a *template*.

    // Check if our serviceName is in the list
    const myTemplate = templates.find((t: any) => t.name === serviceName);
    expect(myTemplate).toBeDefined();
    expect(myTemplate.service_config.command_line_service.env.POSTGRES_URL.plain_text).toBe('postgresql://user:pass@db:5432/testdb');
  });
});
