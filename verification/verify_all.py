
from playwright.sync_api import Page, expect, sync_playwright
import os
import datetime
import time

today = datetime.datetime.now().strftime("%Y-%m-%d")
output_dir = f".audit/ui/{today}"
os.makedirs(output_dir, exist_ok=True)

BASE_URL = "http://localhost:9002"

def verify_dashboard(page: Page):
    print("Navigating to Dashboard...")
    page.goto(BASE_URL + "/")
    page.wait_for_selector("text=Dashboard", timeout=30000)

    print("Verifying Dashboard elements...")
    expect(page.get_by_text("Active Services")).to_be_visible()
    expect(page.get_by_text("Request Volume")).to_be_visible()
    expect(page.get_by_text("Service Health")).to_be_visible()

    screenshot_path = f"{output_dir}/Dashboard.png"
    print(f"Taking screenshot: {screenshot_path}")
    page.screenshot(path=screenshot_path, full_page=True)

def verify_services(page: Page):
    print("Navigating to Services...")
    page.goto(BASE_URL + "/services")
    page.wait_for_selector("text=Services", timeout=30000)

    print("Verifying Services elements...")
    expect(page.get_by_role("button", name="Add Service")).to_be_visible()

    screenshot_path = f"{output_dir}/Services.png"
    print(f"Taking screenshot: {screenshot_path}")
    page.screenshot(path=screenshot_path, full_page=True)

def verify_tools(page: Page):
    print("Navigating to Tools...")
    page.goto(BASE_URL + "/tools")
    page.wait_for_selector("text=Tools", timeout=30000)

    print("Verifying Tools elements...")
    expect(page.get_by_text("Available Tools")).to_be_visible()

    screenshot_path = f"{output_dir}/Tools.png"
    print(f"Taking screenshot: {screenshot_path}")
    page.screenshot(path=screenshot_path, full_page=True)

def verify_resources(page: Page):
    print("Navigating to Resources...")
    page.goto(BASE_URL + "/resources")
    page.wait_for_selector("text=Resources", timeout=30000)

    screenshot_path = f"{output_dir}/Resources.png"
    print(f"Taking screenshot: {screenshot_path}")
    page.screenshot(path=screenshot_path, full_page=True)

def verify_prompts(page: Page):
    print("Navigating to Prompts...")
    page.goto(BASE_URL + "/prompts")
    page.wait_for_selector("text=Prompts", timeout=30000)

    screenshot_path = f"{output_dir}/Prompts.png"
    print(f"Taking screenshot: {screenshot_path}")
    page.screenshot(path=screenshot_path, full_page=True)

def verify_profiles(page: Page):
    print("Navigating to Profiles...")
    page.goto(BASE_URL + "/profiles")
    page.wait_for_selector("text=Profiles", timeout=30000)

    screenshot_path = f"{output_dir}/Profiles.png"
    print(f"Taking screenshot: {screenshot_path}")
    page.screenshot(path=screenshot_path, full_page=True)

def verify_middleware(page: Page):
    print("Navigating to Middleware...")
    page.goto(BASE_URL + "/middleware")
    page.wait_for_selector("text=Middleware", timeout=30000)

    screenshot_path = f"{output_dir}/Middleware.png"
    print(f"Taking screenshot: {screenshot_path}")
    page.screenshot(path=screenshot_path, full_page=True)

def verify_webhooks(page: Page):
    print("Navigating to Webhooks...")
    page.goto(BASE_URL + "/settings/webhooks")
    page.wait_for_selector("text=Webhooks", timeout=30000)

    screenshot_path = f"{output_dir}/Webhooks.png"
    print(f"Taking screenshot: {screenshot_path}")
    page.screenshot(path=screenshot_path, full_page=True)


if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        # Emulate a desktop viewport
        context = browser.new_context(viewport={"width": 1440, "height": 900})
        page = context.new_page()
        try:
            verify_dashboard(page)
            verify_services(page)
            verify_tools(page)
            verify_resources(page)
            verify_prompts(page)
            verify_profiles(page)
            verify_middleware(page)
            verify_webhooks(page)
            print("All verifications passed!")
        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path=f"{output_dir}/Failure.png")
            raise
        finally:
            browser.close()
