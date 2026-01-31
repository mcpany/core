# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import sync_playwright

def verify_stacks():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        # Use a context to inject the API Key if needed, though for public pages/local dev it might be fine.
        # But our client.ts checks for localStorage or assumes valid session.
        # However, for SSR/Client fetch, we might need to set it.
        # Let's see if we need it. The test setup injected X-API-Key.
        # Here we are just visiting the page.
        # If client.ts fails to fetch, it will show an error toast.

        # We can inject localStorage via add_init_script if needed.
        # But let's try without first.

        context = browser.new_context(
            extra_http_headers={'X-API-Key': 'test-token'}
        )

        page = context.new_page()

        print("Navigating to Stacks page...")
        page.goto("http://localhost:9002/stacks")

        print("Waiting for 'Create Stack' button...")
        page.get_by_role("button", name="Create Stack").wait_for(timeout=10000)

        print("Taking screenshot...")
        page.screenshot(path="verification/stacks_page.png")
        print("Screenshot saved to verification/stacks_page.png")

        browser.close()

if __name__ == "__main__":
    verify_stacks()
