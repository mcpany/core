/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './test-data';

test.describe('Tool Wizard E2E', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        // We rely on server/config.minimal.yaml having "complex-tool-service-file" seeded
        await seedUser(request, "wizard-user");

        await page.goto('/login');
        await page.waitForLoadState('networkidle');
        await page.fill('input[name="username"]', 'wizard-user');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupUser(request, "wizard-user");
    });

    test('Should fill out tool wizard and execute', async ({ page }) => {
        await page.goto('/tools');
        // Filter by the specific service loaded from config
        await expect(page.locator('tr', { hasText: 'Complex Tool Service Config' })).toBeVisible();

        // Open Inspector
        await page.locator('tr', { hasText: 'Complex Tool Service Config' }).getByRole('button', { name: 'Inspect' }).click();

        // Check if Wizard is present and selected
        await expect(page.getByRole('tab', { name: 'Wizard' })).toHaveAttribute('data-state', 'active');

        // Debug: Check Schema
        await page.getByRole('tab', { name: 'JSON' }).first().click(); // Schema JSON tab
        const schemaJson = await page.locator('pre').first().textContent();
        console.log("Schema JSON:", schemaJson);
        await page.getByRole('tab', { name: 'Visual' }).click(); // Switch back to Visual

        await expect(page.getByLabel('username')).toBeVisible();

        // Fill Form
        await page.getByLabel('username').fill('johndoe');
        await page.getByLabel('age').fill('30');

        // Select Role (Enum)
        // Select trigger in Shadcn usually has the placeholder or value
        await page.locator('button[role="combobox"]').click();
        await page.getByRole('option', { name: 'admin' }).click();

        // Toggle Active
        // The switch label might be "isActive"
        await page.getByLabel('isActive').click();

        // Add Array Item (Tags)
        await page.getByRole('button', { name: 'Add Item' }).click();
        // The array item input is dynamically generated. It should be inside the array container.
        // We can find it by being the last input in the form or specific scope.
        // Since it's a string array, it renders an Input.
        await page.locator('input[placeholder=""]').last().fill('developer');

        // Nested Object (Address)
        await page.getByLabel('street').fill('123 Main St');
        await page.getByLabel('city').fill('Metropolis');

        // Switch to JSON view to verify
        await page.getByRole('tab', { name: 'JSON' }).click();

        const jsonContent = await page.locator('textarea#args').inputValue();
        const parsed = JSON.parse(jsonContent);

        expect(parsed.username).toBe('johndoe');
        expect(parsed.age).toBe(30);
        expect(parsed.role).toBe('admin');
        expect(parsed.isActive).toBe(true); // default false (or undefined -> false in schema form), clicked once -> true
        expect(parsed.tags).toContain('developer');
        expect(parsed.address.street).toBe('123 Main St');
        expect(parsed.address.city).toBe('Metropolis');

        // Execute
        await page.getByRole('button', { name: 'Execute' }).click();

        // Wait for result
        // The result will be displayed in the Result section
        await expect(page.getByText('Result')).toBeVisible();

        // Since command is "echo", it echoes the args.
        // The tool execution result content should contain our username.
        const resultText = await page.locator('pre').last().textContent();
        expect(resultText).toContain('johndoe');
        expect(resultText).toContain('Metropolis');
    });
});
