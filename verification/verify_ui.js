/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


/**
 * Verification script for the UI.
 * Captures screenshots of key pages (Dashboard, Services, Middleware) to verify visual integrity.
 */
const { chromium } = require('playwright');
const fs = require('fs');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  // Connect to the local dev server
  await page.goto('http://localhost:3000');

  // Verify Dashboard
  await page.waitForTimeout(1000); // Wait for animations/load
  await page.screenshot({ path: '/home/jules/verification/dashboard.png', fullPage: true });

  // Verify Services
  await page.goto('http://localhost:3000/services');
  await page.waitForTimeout(1000);
  await page.screenshot({ path: '/home/jules/verification/services.png', fullPage: true });

  // Verify Middleware
  await page.goto('http://localhost:3000/middleware');
  await page.waitForTimeout(1000);
  await page.screenshot({ path: '/home/jules/verification/middleware.png', fullPage: true });

  await browser.close();
})();
