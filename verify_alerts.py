# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


import time
from playwright.sync_api import sync_playwright, expect

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={'width': 1280, 'height': 800})
        page = context.new_page()

        try:
            print("Navigating to /alerts...")
            page.goto("http://localhost:3000/alerts")

            # Wait for page to load
            time.sleep(2)

            print("Checking page title...")
            # expect(page).to_have_title("Alerts & Incidents") # Title might be different or dynamic

            print("Taking screenshot of Alerts page...")
            page.screenshot(path="/home/jules/verification/alerts_page.png")

            print("Clicking New Alert Rule...")
            page.get_by_role("button", name="New Alert Rule").click()

            time.sleep(1)
            print("Checking dialog fields...")
            expect(page.get_by_label("Metric")).to_be_visible()
            expect(page.get_by_label("Duration")).to_be_visible()

            print("Filling form...")
            page.get_by_label("Name").fill("Test Rule")
            page.get_by_label("Metric").fill("cpu_usage")
            page.get_by_placeholder("Threshold").fill("90")
            page.get_by_label("Duration").fill("5m")

            print("Taking screenshot of Create Rule Dialog...")
            page.screenshot(path="/home/jules/verification/create_rule_dialog.png")

            # Navigate to Settings
            print("Navigating to /settings...")
            page.goto("http://localhost:3000/settings")
            time.sleep(2)

            print("Checking Health Alerts settings...")
            expect(page.get_by_text("Health Alerts")).to_be_visible()
            expect(page.get_by_label("Webhook URL")).to_be_visible()

            print("Taking screenshot of Settings page...")
            page.screenshot(path="/home/jules/verification/settings_page.png")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="/home/jules/verification/error.png")
            raise e
        finally:
            browser.close()

if __name__ == "__main__":
    run()
