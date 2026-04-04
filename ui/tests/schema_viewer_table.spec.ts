/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, request } from '@playwright/test';

test.describe('Schema Viewer Table View', () => {
  const serviceName = 'e2e-schema-test-service';

  test.beforeAll(async () => {
    // Seed the database with a test service containing an OpenAPI tool with a complex schema
    const apiContext = await request.newContext();
    const response = await apiContext.post('http://localhost:8080/api/v1/services', {
      data: {
        name: serviceName,
        openapi_service: {
            address: "http://example.com",
            tools: [
                {
                    name: "complex_tool",
                    description: "A tool with a complex nested schema",
                    inputSchema: {
                        type: "object",
                        properties: {
                            user_id: {
                                type: "string",
                                description: "The ID of the user"
                            },
                            preferences: {
                                type: "object",
                                description: "User preferences",
                                properties: {
                                    theme: {
                                        type: "string",
                                        enum: ["light", "dark"],
                                        default: "light"
                                    },
                                    notifications: {
                                        type: "boolean",
                                        description: "Enable notifications"
                                    }
                                }
                            },
                            tags: {
                                type: "array",
                                items: {
                                    type: "string"
                                },
                                description: "List of tags"
                            }
                        },
                        required: ["user_id"]
                    }
                }
            ]
        },
        priority: 15
      }
    });
    expect(response.ok()).toBeTruthy();
    await apiContext.dispose();
  });

  test.afterAll(async () => {
    // Clean up
    const apiContext = await request.newContext();
    await apiContext.delete(`http://localhost:8080/api/v1/services/${serviceName}`);
    await apiContext.dispose();
  });

  test('should render complex schema in a formatted table in the Discovered Tools modal', async ({ page }) => {
    // Navigate to the service detail page
    await page.goto(`/upstream-services/${serviceName}`);

    // Wait for page to load
    await expect(page.getByRole('heading', { level: 1 })).toContainText(serviceName);

    // Go to Settings tab
    await page.getByRole('tab', { name: 'Settings' }).click();

    // Click "Validate" to trigger the validation endpoint and discover tools
    const validateButton = page.getByRole('button', { name: 'Validate' });
    await validateButton.click();

    // Wait for the Discovered Capabilities dialog to appear
    await expect(page.getByRole('dialog', { name: 'Discovered Capabilities' })).toBeVisible();

    // Expand the "complex_tool" accordion item
    const toolAccordion = page.getByRole('button', { name: /complex_tool/i });
    await toolAccordion.click();

    // Verify the schema table headers
    await expect(page.getByRole('cell', { name: 'Property' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'Type' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'Description' })).toBeVisible();

    // Verify nested properties are visible in the table
    await expect(page.getByText('user_id')).toBeVisible();
    await expect(page.getByText('The ID of the user')).toBeVisible();

    await expect(page.getByText('preferences')).toBeVisible();
    await expect(page.getByText('User preferences')).toBeVisible();

    await expect(page.getByText('theme')).toBeVisible();
    await expect(page.getByText('light, dark')).toBeVisible(); // Enum values

    await expect(page.getByText('tags')).toBeVisible();
    await expect(page.getByText('List of tags')).toBeVisible();

    // Close the dialog
    await page.getByRole('button', { name: 'Done' }).click();
  });
});
