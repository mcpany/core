/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('UI Crawler', () => {
  test('should crawl and verify all internal links', async ({ page }) => {
    const visited = new Set<string>();
    const toVisit: string[] = ['/'];
    const failures: { url: string; error: string }[] = [];
    const maxPages = 50;

    // Helper to normalize URLs (strip query params and hashes for deduplication)
    const normalize = (url: string) => {
      try {
         // Handle relative URLs by using a dummy base
         const u = new URL(url, 'http://localhost');
         return u.pathname;
      } catch {
         return url;
      }
    };

    while (toVisit.length > 0 && visited.size < maxPages) {
      const url = toVisit.shift();
      if (!url) break;

      const normalizedUrl = normalize(url);
      if (visited.has(normalizedUrl)) continue;
      visited.add(normalizedUrl);

      console.log(`Visiting: ${url}`);

      try {
        const response = await page.goto(url, { waitUntil: 'domcontentloaded' });

        if (!response) {
            failures.push({ url, error: 'No response received' });
            continue;
        }

        if (response.status() >= 400) {
          failures.push({ url, error: `Status ${response.status()}` });
        }

        // Collect new links
        const anchors = await page.locator('a[href]').all();
        for (const anchor of anchors) {
          const href = await anchor.getAttribute('href');
          if (href && href.startsWith('/') && !href.startsWith('//') && !visited.has(normalize(href))) {
             if (!toVisit.includes(href) && !href.startsWith('/api/')) {
                 toVisit.push(href);
             }
          }
        }

      } catch (e: unknown) {
        // Ignore navigation rejections if they are just redirects or similar benign issues?
        // Actually, we want to catch errors.
        const errorMessage = e instanceof Error ? e.message : String(e);
        failures.push({ url, error: errorMessage });
      }
    }

    if (failures.length > 0) {
      console.error('Crawler Failures:', failures);
    }
    expect(failures, `Failed pages: ${JSON.stringify(failures, null, 2)}`).toHaveLength(0);
  });
});
