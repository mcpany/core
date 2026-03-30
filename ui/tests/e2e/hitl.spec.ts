import { test, expect } from '@playwright/test';

// In this E2E test, we must use real interactions. The server exposes a seeded database or state.
// From `api_hitl.go` init():
// globalHITLState.approvals["1"] = database.drop_table (MFA=true)
// globalHITLState.approvals["2"] = aws.terminate_instance (MFA=false)

test.describe('HITL Dashboard flow', () => {
    test.beforeEach(async ({ page }) => {
        // Go to the HITL dashboard
        await page.goto('/hitl');
    });

    test('should show pending approvals from server seed', async ({ page }) => {
        await expect(page.getByText('database.drop_table')).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('aws.terminate_instance')).toBeVisible();
    });

    test('should allow denying an approval without MFA', async ({ page }) => {
        // Find the card for aws.terminate_instance
        const awsCard = page.locator('.rounded-xl, .border, .bg-card').filter({ hasText: 'aws.terminate_instance' }).first();
        await expect(awsCard).toBeVisible();

        // Click Deny
        await awsCard.getByRole('button', { name: 'Deny' }).click();

        // The card should disappear or change status. Since the backend deletes it:
        await expect(page.getByText('aws.terminate_instance')).toBeHidden({ timeout: 5000 });
    });

    test('should prompt for MFA and allow approval for sensitive tool', async ({ page }) => {
        // Find the card for database.drop_table
        const dbCard = page.locator('.rounded-xl, .border, .bg-card').filter({ hasText: 'database.drop_table' }).first();
        await expect(dbCard).toBeVisible();

        // Click Approve
        await dbCard.getByRole('button', { name: 'Approve' }).click();

        // Expect MFA Dialog
        const dialog = page.locator('[role="dialog"]');
        await expect(dialog.getByText('Multi-Factor Authentication Required')).toBeVisible();

        // Enter MFA code
        await dialog.getByPlaceholder('MFA Code').fill('987654');

        // Verify & Approve
        await dialog.getByRole('button', { name: 'Verify & Approve' }).click();

        // The dialog should close
        await expect(dialog).toBeHidden();

        // The card should disappear
        await expect(page.getByText('database.drop_table')).toBeHidden({ timeout: 5000 });
    });
});
