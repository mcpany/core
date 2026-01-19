
import { test, expect } from '@playwright/test';

test.describe('Tools Page', () => {
    test.beforeEach(async ({ page }) => {
        // Mock tools API
        await page.route('**/api/v1/tools', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    tools: [
                        { name: 'tool1', description: 'Tool 1', serviceId: 'service1', disable: false },
                        { name: 'tool2', description: 'Tool 2', serviceId: 'service2', disable: false },
                        { name: 'tool3', description: 'Tool 3', serviceId: 'service1', disable: true },
                    ]
                })
            });
        });

        // Mock services API
        await page.route('**/api/v1/services', async (route) => {
             // Return array format (or { services: [] } if needed, but client handles both)
             // The client code (ui/src/lib/client.ts) handles { services: [] } OR []
             // Let's verify what the client expects. It handles both.
             await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    services: [
                        { id: 'service1', name: 'Service 1' },
                        { id: 'service2', name: 'Service 2' }
                    ]
                })
            });
        });

        // Mock doctor API to prevent system status banner
        await page.route('**/doctor', async route => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ status: 'healthy', checks: {} })
            });
        });
    });

    test('should filter tools by service', async ({ page }) => {
        await page.goto('/tools');

        // Verify initial state (All Services)
        await expect(page.getByText('tool1')).toBeVisible();
        await expect(page.getByText('tool2')).toBeVisible();
        await expect(page.getByText('tool3')).toBeVisible();

        // Select Service 1
        // Find the select trigger. We look for "Filter by Service" placeholder text first
        // If a value is selected, the text changes. Initially "All Services" might be selected if default?
        // Code sets default to "all".
        // Trigger displays value.
        // The trigger initially shows "Filter by Service" because value "all" doesn't have a label in the trigger
        // UNLESS the SelectValue has placeholder.
        // Wait, SelectValue renders the selected item's children?
        // If "all" is selected, does "All Services" show?
        // Let's check the code:
        // <SelectItem value="all">All Services</SelectItem>
        // Yes, initially "All Services" should be displayed if "all" is default.
        // But the placeholder is "Filter by Service".
        // If value is set, placeholder is hidden?

        // Let's try to click the trigger. It is a button with role combobox usually.
        // But let's use getByRole('combobox').
        const filterTrigger = page.getByRole('combobox');
        await expect(filterTrigger).toBeVisible();
        await filterTrigger.click();

        // Select Service 1
        await page.getByRole('option', { name: 'Service 1' }).click();

        // Verify filtered list
        await expect(page.getByText('tool1')).toBeVisible();
        await expect(page.getByText('tool3')).toBeVisible();
        await expect(page.getByText('tool2')).not.toBeVisible();

        // Select Service 2
        await filterTrigger.click();
        await page.getByRole('option', { name: 'Service 2' }).click();

        // Verify filtered list
        await expect(page.getByText('tool2')).toBeVisible();
        await expect(page.getByText('tool1')).not.toBeVisible();
        await expect(page.getByText('tool3')).not.toBeVisible();

        // Select All Services
        await filterTrigger.click();
        await page.getByRole('option', { name: 'All Services' }).click();

        // Verify all visible again
        await expect(page.getByText('tool1')).toBeVisible();
        await expect(page.getByText('tool2')).toBeVisible();
    });
});
