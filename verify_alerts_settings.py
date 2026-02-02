# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


import time
from playwright.sync_api import sync_playwright, expect

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={'width': 1280, 'height': 1200}) # Increased height
        page = context.new_page()

        try:
            # ... (previous alert steps skipped for speed if verified, but keeping for completeness)

            # Navigate to Settings directly to verify scrolling
            print("Navigating to /settings...")
            page.goto("http://localhost:3000/settings")
            time.sleep(2)

            print("Checking Health Alerts settings...")
            # Scroll to bottom
            page.evaluate("window.scrollTo(0, document.body.scrollHeight)")
            time.sleep(1)

            health_alerts = page.get_by_text("Health Alerts")
            expect(health_alerts).to_be_visible()
            expect(page.get_by_label("Webhook URL")).to_be_visible()

            # Take screenshot of the Health Alerts section
            # Ideally locate the container
            print("Taking screenshot of Settings page (bottom)...")
            page.screenshot(path="/home/jules/verification/settings_page_full.png", full_page=True)

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="/home/jules/verification/error.png")
            raise e
        finally:
            browser.close()

if __name__ == "__main__":
    run()
