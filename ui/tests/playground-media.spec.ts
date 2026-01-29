/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Playground Rich Media', () => {
  const serviceId = 'e2e_media_service';

  test.beforeAll(async ({ request }) => {
    // Register a command_line service that echoes an image
    // We use a small 1x1 pixel PNG (red color)
    const base64Image = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==";

    // We need to craft a JSON string that the shell command will echo.
    // The command is: echo '{"content": [{"type": "image", "data": "...", "mimeType": "image/png"}]}'
    // We need to be careful with escaping quotes for the shell.
    const jsonOutput = JSON.stringify({
        content: [
            {
                type: "image",
                data: base64Image,
                mimeType: "image/png"
            },
            {
                type: "text",
                text: "# Markdown Test\n\nThis is a **markdown** test."
            }
        ]
    });

    // Escape for shell: echo 'JSON' -> echo '{"key": "val"}'
    // We need to escape backslashes because some 'echo' implementations interpret them.
    // JSON.stringify gives us literal backslashes (e.g. for \n). We want echo to output them literally.
    // So we double them.
    const escapedJsonOutput = jsonOutput.replace(/\\/g, '\\\\').replace(/'/g, "'\\''");

    const payload = {
        id: serviceId,
        name: serviceId,
        command_line_service: {
            command: "sh",
            calls: {
                "echo_call": {
                    // We pass the command string to sh -c
                    args: ["-c", `echo '${escapedJsonOutput}'`]
                }
            },
            tools: [
                {
                    name: "echo_media",
                    description: "Returns an image and markdown",
                    call_id: "echo_call",
                    input_schema: {
                        type: "object",
                        properties: {}
                    }
                }
            ]
        }
    };

    const res = await request.post('/api/v1/services', {
        data: payload
    });

    if (!res.ok()) {
        console.error('Failed to register service:', await res.text());
        throw new Error('Failed to register e2e_media_service');
    }
  });

  test.afterAll(async ({ request }) => {
    await request.delete(`/api/v1/services/${serviceId}`);
  });

  test('should render images and markdown in playground', async ({ page }) => {
    await page.goto('/playground');

    // Wait for tools to load
    // Open sidebar if closed (it's open by default on desktop, but let's ensure)
    // Actually, PlaygroundClientPro loads tools on mount.

    // Check if echo_media is in the list
    // It might take a moment for the service to be picked up?
    // We can try to reload or wait.

    // To be safe, let's search for it.
    await page.getByPlaceholder('Enter command or select a tool...').fill('echo_media');

    // It should appear in suggestions
    await expect(page.getByText('echo_media')).toBeVisible();

    // Click it to select
    await page.getByText('echo_media').click();

    // Dialog opens. Click "Build Command" (or Submit)
    // The dialog button usually says "Run Tool" or similar.
    // In PlaygroundClientPro: ToolForm onSubmit -> calls handleToolFormSubmit -> sets input -> calls processResponse (actually just sets input and user hits enter, OR auto-runs?
    // Let's check PlaygroundClientPro logic again.
    // handleToolFormSubmit: setToolToConfigure(null); setInput(command);
    // It does NOT auto run. It just populates the input.

    await page.getByRole('button', { name: /build command/i }).click();

    // Wait for the input to be populated
    await expect(page.getByRole('textbox', { name: /enter command/i })).toHaveValue(/echo_media/);

    // Click Send
    await page.getByRole('button', { name: 'Send' }).click();

    // Wait for result
    // We expect an image
    await expect(page.locator('img[alt="Tool Output"]')).toBeVisible();

    // Verify Markdown rendering (H1)
    await expect(page.getByRole('heading', { name: 'Markdown Test' })).toBeVisible();

    // Verify "Rich" button is active/visible
    await expect(page.getByRole('button', { name: 'Rich' })).toBeVisible();
  });
});
