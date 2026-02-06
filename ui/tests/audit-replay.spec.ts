
import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';
import * as path from 'path';

test.describe('Audit Replay Flow', () => {
  test.beforeAll(() => {
    console.log('Seeding audit logs...');
    try {
        // Run from ui/ directory context
        const scriptPath = path.join(__dirname, '../scripts/seed_audit_logs.ts');
        execSync(`npx tsx ${scriptPath}`, { stdio: 'inherit' });
    } catch (e) {
        console.error('Failed to seed logs:', e);
    }
  });

  test('should inspect log and replay in playground', async ({ page }) => {
    // Go to Audit Logs page
    await page.goto('/audit');

    // Wait for table to load
    await expect(page.locator('table')).toBeVisible();

    // Reload to ensure new logs are picked up
    await page.reload();
    // Wait for loading spinner to disappear if any, or just wait for table rows
    await expect(page.locator('table')).toBeVisible();
    await page.waitForTimeout(2000); // Allow fetch to complete

    // Check if "No logs found" is displayed
    const noLogs = await page.getByText('No logs found').isVisible();
    if (noLogs) {
        console.warn('No logs found. Seeding might have failed or backend is down.');
        // Fail the test if no logs, because we expect seeding to work
        throw new Error('No logs found after seeding.');
    }

    // Click the first "View" button
    // Use more specific selector to avoid other view buttons if any
    await page.locator('table button').filter({ hasText: 'View' }).first().click();

    // Check for Sheet (it renders as a dialog role)
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByText('Arguments')).toBeVisible();

    // Check for Monaco Editor presence
    // Monaco adds a class 'monaco-editor'
    await expect(page.locator('.monaco-editor').first()).toBeVisible({ timeout: 15000 });

    // Click Replay
    await page.getByRole('button', { name: 'Replay' }).click();

    // Assert navigation to Playground
    await expect(page).toHaveURL(/\/playground\?tool=.*&args=.*/);

    // Verify Input in Playground is populated
    const input = page.locator('input[placeholder="Enter command or select a tool..."]');
    await expect(input).toBeVisible();

    // Wait for value to be populated
    await expect(async () => {
        const val = await input.inputValue();
        expect(val).not.toBe('');
    }).toPass();
  });
});
