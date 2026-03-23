import { test, expect } from '@playwright/test';
import { seedGlobalState } from './test-data';

test.describe('Skills Bulk Actions', () => {
    test.beforeAll(async () => {
        // Seed the database to ensure a clean state
        await seedGlobalState();
    });

    test.beforeEach(async ({ page }) => {
        // Log in
        await page.goto('/login');
        await page.fill('input[placeholder="Username"]', 'admin');
        await page.fill('input[type="password"]', 'admin');
        await page.click('button:has-text("Login")');
        await page.waitForURL('/');
    });

    test('should allow bulk selection and deletion of skills', async ({ page }) => {
        await page.goto('/skills');

        // Create first skill
        await page.click('text="Create Skill"');
        await page.fill('input[name="name"]', 'skill_to_bulk_delete_1');
        await page.fill('input[name="description"]', 'Desc 1');
        await page.click('button:has-text("Save")');
        await expect(page.locator('h1', { hasText: 'Agent Skills' })).toBeVisible();

        // Create second skill
        await page.click('text="Create Skill"');
        await page.fill('input[name="name"]', 'skill_to_bulk_delete_2');
        await page.fill('input[name="description"]', 'Desc 2');
        await page.click('button:has-text("Save")');
        await expect(page.locator('h1', { hasText: 'Agent Skills' })).toBeVisible();

        // Verify skills are visible
        await expect(page.locator('text="skill_to_bulk_delete_1"')).toBeVisible();
        await expect(page.locator('text="skill_to_bulk_delete_2"')).toBeVisible();

        // Select the first skill
        await page.click('button:has-text("Select All")');
        // Now "Deselect All" button should be visible, meaning selection worked.
        await expect(page.locator('button:has-text("Deselect All")')).toBeVisible();
        // The Bulk Delete button should be visible
        await expect(page.locator('button:has-text("Bulk Delete")')).toBeVisible();

        // Ensure both are selected by counting selected text or just hitting bulk delete.
        // Accept the confirm dialog
        page.on('dialog', dialog => dialog.accept());
        await page.click('button:has-text("Bulk Delete")');

        // Wait for the deletion toast to indicate it happened.
        await expect(page.locator('text="Successfully deleted"')).toBeVisible();

        // Verify skills are removed from the list
        await expect(page.locator('text="skill_to_bulk_delete_1"')).toBeHidden();
        await expect(page.locator('text="skill_to_bulk_delete_2"')).toBeHidden();
    });
});
