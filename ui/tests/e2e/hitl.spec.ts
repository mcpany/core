import { test, expect } from '@playwright/test';

test('HITL Approvals Dashboard', async ({ page }) => {
    // Navigate to HITL page
    await page.goto('/hitl');

    // Verify page title
    await expect(page.locator('h1')).toHaveText('HITL Approvals');

    // Wait for data to load
    await expect(page.locator('.lucide-loader2')).not.toBeVisible();

    // Verify seeded data (from server init)
    // We expect "database.drop_table" and "aws.terminate_instance"
    await expect(page.locator('text=database.drop_table')).toBeVisible();
    await expect(page.locator('text=aws.terminate_instance')).toBeVisible();

    // Find the aws.terminate_instance card and click Approve
    const awsCard = page.locator('[data-testid="hitl-card"]').filter({ hasText: 'aws.terminate_instance' });
    await awsCard.locator('button:has-text("Approve")').click();

    // Verify toast notification (exact text to avoid matching screen reader text)
    await expect(page.getByText('Action Approved', { exact: true }).first()).toBeVisible();
});
