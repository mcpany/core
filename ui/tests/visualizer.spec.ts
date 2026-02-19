/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser } from './e2e/test-data';

const BASE_URL = process.env.BACKEND_URL || 'http://localhost:50050';
const API_KEY = process.env.MCPANY_API_KEY || 'test-token';
const HEADERS = { 'X-API-Key': API_KEY };

test.describe('Visualizer E2E', () => {
  test.beforeEach(async ({ request, page }) => {
    await seedServices(request);
    await seedUser(request, "vis-admin");

    // Login
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'vis-admin');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
  });

  test.afterEach(async ({ request }) => {
    await cleanupServices(request);
  });

  test('should visualize agent flow after tool execution', async ({ page, request }) => {
    // 1. Find the Echo Tool name
    const toolsRes = await request.get('/api/v1/tools', { headers: HEADERS });
    const toolsData = await toolsRes.json();

    // Handle both array and object response formats
    const tools = Array.isArray(toolsData) ? toolsData : (toolsData.tools || []);
    const echoTool = tools.find((t: any) => t.name.includes('echo_tool'));

    if (!echoTool) {
        console.log('Tools available:', JSON.stringify(tools, null, 2));
        throw new Error('Echo tool not found in seeded services');
    }

    const toolName = echoTool.name;
    console.log(`Found tool: ${toolName}`);

    // 2. Execute a tool to generate a trace
    const response = await request.post('/api/v1/execute', {
        data: {
            name: toolName,
            arguments: { input: "hello" }
        },
        headers: HEADERS
    });

    expect(response.ok()).toBeTruthy();

    // 3. Navigate to Visualizer
    await page.goto('/visualizer');

    // 4. Wait for polling to pick up the trace (interval is 3s)
    // We check for the presence of nodes.
    // The graph should contain "Client" (UserNode) and "MCP Core" (AgentNode).
    // Note: CLI tool execution might not appear as a separate span if tracing only captures HTTP traffic.

    // Check for Client node
    await expect(page.locator('text=Client')).toBeVisible({ timeout: 10000 });

    // Check for MCP Core node
    await expect(page.locator('text=MCP Core')).toBeVisible({ timeout: 10000 });

    // Check that we have a graph (canvas exists)
    await expect(page.locator('.react-flow__pane')).toBeVisible();
  });
});
