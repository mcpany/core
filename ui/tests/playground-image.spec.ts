import { test, expect } from '@playwright/test';

test.describe('Playground Image Rendering', () => {
    test('should render an image returned by a tool', async ({ page }) => {
        // Mock the tool list to include a tool that returns an image
        await page.route('**/api/v1/tools', async route => {
            await route.fulfill({
                json: {
                    tools: [
                        {
                            name: 'generate_image',
                            description: 'Generates a test image',
                            inputSchema: {
                                type: 'object',
                                properties: {}
                            }
                        }
                    ]
                }
            });
        });

        // Mock the execute endpoint to return a CallToolResult with an image
        // This simulates a real backend response for a tool that returns an image content item
        await page.route('**/api/v1/execute', async route => {
            const base64Data = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=';
            await route.fulfill({
                json: {
                    content: [
                        {
                            type: 'image',
                            data: base64Data,
                            mimeType: 'image/png'
                        }
                    ],
                    isError: false
                }
            });
        });

        await page.goto('/playground');

        // Wait for the tool to appear in the sidebar
        await expect(page.getByText('generate_image')).toBeVisible();

        // Click the tool to select it
        await page.getByText('generate_image').click();

        // Wait for configuration dialog
        await expect(page.getByRole('dialog')).toBeVisible();
        await expect(page.getByRole('heading', { name: 'generate_image' })).toBeVisible();

        // Click "Build Command" (or whatever the submit button is)
        // In ToolForm, the button says "Build Command" or similar.
        // Let's target by type submit
        await page.getByRole('button', { name: /Build Command/i }).click();

        // Now we should be back in the console, with the input populated
        await expect(page.getByRole('textbox', { name: /enter command/i })).toHaveValue(/generate_image/);

        // Click Send
        await page.getByLabel('Send').click();

        // Verify the chat message contains the image
        // The SmartResultRenderer should render an <img> tag with the base64 data
        const imgSelector = 'img[src^="data:image/png;base64"]';
        const img = page.locator(imgSelector);

        await expect(img).toBeVisible();
        await expect(img).toHaveAttribute('src', expect.stringContaining('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII='));

        // Verify the "Visual" label/button is present (indicating Smart View is active)
        await expect(page.getByText('Visual')).toBeVisible();
    });
});
