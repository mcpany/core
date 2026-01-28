/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Smart Result View', () => {
    test.beforeAll(async ({ request }) => {
        // Register a service that returns a JSON list via echo
        // We use relative path so it works with PLAYWRIGHT_BASE_URL in K8s environment (proxied via UI)
        // or local environment.

        // Register the service
        const response = await request.post(`/api/v1/services`, {
            data: {
                id: "test-table-service",
                name: "test-table-service",
                command_line_service: {
                    tools: [
                        {
                            name: "get_test_table",
                            description: "Returns a JSON list for table testing",
                            call_id: "echo_json"
                        }
                    ],
                    calls: {
                        echo_json: {
                            // We use 'echo' which is standard in most linux distros (including alpine)
                            command: "echo",
                            // Return a JSON string representing a list of objects
                            // Note: Single quotes for shell argument, double quotes for JSON
                            args: ['[{"id": 1, "name": "Alice", "role": "Admin"}, {"id": 2, "name": "Bob", "role": "User"}]']
                        }
                    }
                }
            }
        });

        // If it fails, log why
        if (!response.ok()) {
            console.error('Failed to register service:', await response.text());
        }
        expect(response.ok()).toBeTruthy();
    });

    test('should render JSON list as a table and allow toggling to raw view', async ({ page }) => {
        await page.goto('/playground');

        // Execute the tool
        // We type the command directly into the input
        await page.getByPlaceholder('Enter command or select a tool...').fill('get_test_table {}');
        await page.getByLabel('Send').click();

        // Wait for result
        // The table should appear.
        // We look for column headers.
        // Note: shadcn Table headers are th elements with generic roles, but Playwright's getByRole('columnheader') matches th.
        // We verify the "Result: get_test_table" header is present to ensure we are looking at the right message
        await expect(page.getByText('Result: get_test_table')).toBeVisible();

        // Check for Smart View buttons (Table/JSON)
        await expect(page.getByRole('button', { name: 'Table' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'JSON' })).toBeVisible();

        // Verify Table headers
        await expect(page.getByRole('columnheader', { name: 'id', exact: true })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'name', exact: true })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'role', exact: true })).toBeVisible();

        // Check rows content
        await expect(page.getByRole('cell', { name: 'Alice' })).toBeVisible();
        await expect(page.getByRole('cell', { name: 'Bob' })).toBeVisible();

        // Toggle to Raw View
        await page.getByRole('button', { name: 'JSON' }).click();

        // Table headers should be gone
        await expect(page.getByRole('columnheader', { name: 'id', exact: true })).toBeHidden();

        // Raw JSON should be visible
        // The raw result from echo command contains "stdout" field
        await expect(page.getByText('"stdout":')).toBeVisible();
        // And the escaped JSON string content
        await expect(page.getByText('Alice')).toBeVisible();

        // Toggle back to Smart View (Table)
        await page.getByRole('button', { name: 'Table' }).click();
        await expect(page.getByRole('columnheader', { name: 'id', exact: true })).toBeVisible();
    });
});
