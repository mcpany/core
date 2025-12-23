# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


from playwright.sync_api import Page, expect, sync_playwright

def verify_middleware(page: Page):
    print("Navigating to Middleware Settings...")
    page.goto("http://localhost:9002/settings/middleware")

    print("Verifying Middleware elements...")
    expect(page.get_by_text("Middleware Pipeline")).to_be_visible()
    expect(page.get_by_text("Global Rate Limiter")).to_be_visible()

    print("Taking screenshot...")
    page.screenshot(path="verification/middleware_verified.png", full_page=True)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_middleware(page)
        finally:
            browser.close()
