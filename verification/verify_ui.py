# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


import os
import time
from playwright.sync_api import sync_playwright, expect

def verify_dashboard_and_navigation():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        # Use a larger viewport to capture the sidebar and main content nicely
        context = browser.new_context(viewport={"width": 1280, "height": 720})
        page = context.new_page()

        try:
            # 1. Dashboard Load
            print("Navigating to Dashboard...")
            page.goto("http://localhost:3000")

            # Wait for key elements to be visible
            expect(page.get_by_role("heading", name="Dashboard")).to_be_visible(timeout=10000)
            expect(page.get_by_text("Total Requests")).to_be_visible()
            expect(page.get_by_text("System Health")).to_be_visible()
            expect(page.get_by_text("Request Volume")).to_be_visible()

            # Screenshot Dashboard
            print("Taking Dashboard Screenshot...")
            page.screenshot(path="verification/dashboard.png")

            # 2. Services Page
            print("Navigating to Services...")
            page.goto("http://localhost:3000/services")
            expect(page.get_by_role("button", name="Add Service")).to_be_visible()

            # Open Add Service Sheet
            page.get_by_role("button", name="Add Service").click()
            expect(page.get_by_text("New Service")).to_be_visible()

            # Screenshot Services with Sheet Open
            print("Taking Services Screenshot...")
            page.screenshot(path="verification/services_sheet.png")

            # Close sheet
            page.keyboard.press("Escape")

            # 3. Tools Page
            print("Navigating to Tools...")
            page.goto("http://localhost:3000/tools")
            expect(page.get_by_text("Available Tools")).to_be_visible()
            print("Taking Tools Screenshot...")
            page.screenshot(path="verification/tools.png")

            # 4. Resources Page
            print("Navigating to Resources...")
            page.goto("http://localhost:3000/resources")
            expect(page.get_by_text("Managed resources")).to_be_visible()
            print("Taking Resources Screenshot...")
            page.screenshot(path="verification/resources.png")

            # 5. Prompts Page
            print("Navigating to Prompts...")
            page.goto("http://localhost:3000/prompts")
            expect(page.get_by_text("System Prompts")).to_be_visible()
            print("Taking Prompts Screenshot...")
            page.screenshot(path="verification/prompts.png")

            # 6. Middleware Page
            print("Navigating to Middleware...")
            page.goto("http://localhost:3000/middleware")
            expect(page.get_by_text("Middleware Pipeline")).to_be_visible()
            print("Taking Middleware Screenshot...")
            page.screenshot(path="verification/middleware.png")

            # 7. Webhooks Page
            print("Navigating to Webhooks...")
            page.goto("http://localhost:3000/webhooks")
            expect(page.get_by_text("Configured Webhooks")).to_be_visible()
            print("Taking Webhooks Screenshot...")
            page.screenshot(path="verification/webhooks.png")

        except Exception as e:
            print(f"Verification failed: {e}")
            # Take a screenshot on failure to help debug
            page.screenshot(path="verification/failure.png")
            raise e
        finally:
            browser.close()

if __name__ == "__main__":
    verify_dashboard_and_navigation()
