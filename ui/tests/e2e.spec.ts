/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { test, expect } from '@playwright/test';
import path from 'path';

const DATE = new Date().toISOString().split('T')[0];
const AUDIT_DIR = path.join(__dirname, `../../.audit/ui/${DATE}`);

test.describe('MCP Any UI E2E Tests', () => {

  test.beforeEach(async ({ page }) => {
    // Mock services
    await page.route('**/api/services', async (route) => {
      await route.fulfill({
        json: {
          services: [
            {
              id: "svc_01",
              name: "Payment Gateway",
              connection_pool: { max_connections: 100 },
              disable: false,
              version: "v1.2.0",
              http_service: { address: "https://stripe.com", tools: [], resources: [] }
            },
            {
               id: "svc_02",
               name: "User Service",
               disable: false,
               version: "v1.0",
               grpc_service: { address: "localhost:50051", tools: [], resources: [] }
            }
          ]
        }
      });
    });
    // Mock tools for Dashboard/Tools page checks
    await page.route('**/api/tools', async (route) => {
         await route.fulfill({
             json: {
                  tools: [
                       { name: "calculator", description: "calc", source: "discovered", serviceName: "Math" },
                       { name: "weather_lookup", description: "weather", source: "configured", serviceName: "Weather" }
                  ]
             }
         });
    });
    // Mock resources for Dashboard checks (if needed) or explicit page visits
    await page.route('**/api/resources', async (route) => {
         await route.fulfill({ json: [] });
    });
    // Mock dashboard metrics
    await page.route('**/api/dashboard/metrics', async (route) => {
        await route.fulfill({
            json: [
                { label: "Total Requests", value: "1,234", icon: "Activity", change: "+12%", trend: "up" },
                { label: "Active Services", value: "5", icon: "Server", change: "0", trend: "neutral" },
                { label: "System Health", value: "98%", icon: "Zap", change: "+1%", trend: "up" },
                { label: "Error Rate", value: "0.1%", icon: "AlertCircle", change: "-0.5%", trend: "up" }
            ]
        });
    });
  });

  test('Dashboard loads correctly', async ({ page }) => {
    await page.goto('/');
    // Check for metrics
    await expect(page.locator('text=Total Requests')).toBeVisible();
    await expect(page.locator('text=Active Services')).toBeVisible();
    // Check for health widget
    // Use .first() because "System Health" might be in metrics and widget title
    await expect(page.locator('text=System Health').first()).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'dashboard_verified.png'), fullPage: true });
  });

  test('Services page lists services and allows toggle', async ({ page }) => {
    await page.goto('/services');
    await expect(page.locator('h2')).toContainText('Services');

    // Check for list of services
    await expect(page.locator('text=Payment Gateway')).toBeVisible();
    await expect(page.locator('text=User Service')).toBeVisible();


    // Test toggle
    const switchElement = page.locator('button[role="switch"]').first();
    await expect(switchElement).toBeVisible();

    // Test Add Service (Edit Sheet)
    await page.click('text=Add Service');
    await expect(page.locator('text=New Service')).toBeVisible();
    await expect(page.locator('input#name')).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'services_verified.png'), fullPage: true });
  });

  test('Tools page lists tools', async ({ page }) => {
    await page.goto('/tools');
    await expect(page.locator('h2')).toContainText('Tools');
    // Use accessible selectors looking for cell content
    await expect(page.getByRole('cell', { name: 'calculator' })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'weather_lookup' })).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'tools.png'), fullPage: true });
  });

  test('Middleware page shows pipeline', async ({ page }) => {
    await page.goto('/middleware');
    await expect(page.locator('h2')).toContainText('Middleware Pipeline');
    await expect(page.locator('text=Incoming Request')).toBeVisible();
    // Use first() to avoid strict mode violation if text appears multiple times
    await expect(page.locator('text=auth').first()).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'middleware.png'), fullPage: true });
  });

  test('Webhooks page displays configuration', async ({ page }) => {
    await page.goto('/settings/webhooks');
    await expect(page.getByRole('heading', { name: 'Webhooks' })).toBeVisible();

    // Audit Screenshot
    await page.screenshot({ path: path.join(AUDIT_DIR, 'webhooks_verified.png'), fullPage: true });
  });

});
