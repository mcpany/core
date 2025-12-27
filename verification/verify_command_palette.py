
from playwright.sync_api import sync_playwright
import time

def verify_command_palette():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        try:
            # Navigate to the app
            page.goto("http://localhost:9002")

            # Wait for the page to load
            page.wait_for_selector("text=Dashboard", timeout=60000)

            # 1. Test Search Trigger Button (Desktop)
            search_trigger = page.locator(".hidden.md\\:flex.items-center.text-sm.text-muted-foreground.bg-muted\\/50")
            if search_trigger.is_visible():
                print("Search trigger is visible.")
            else:
                print("Search trigger not found or hidden.")

            # 2. Open Command Palette with Keyboard Shortcut
            page.keyboard.press("Meta+k")

            # Wait for dialog
            page.wait_for_selector('div[role="dialog"]', state='visible')

            # Check for input
            expect_input = page.get_by_placeholder("Type a command or search...")
            if expect_input.is_visible():
                print("Command palette input is visible.")

            # Check for some items - using more specific locators
            # cmdk items often have role="option"
            dashboard_item = page.get_by_role("option", name="Dashboard")
            if dashboard_item.is_visible():
                print("Dashboard item visible.")

            dark_mode_item = page.get_by_role("option", name="Dark Mode")
            if dark_mode_item.is_visible():
                print("Dark Mode item visible.")

            # Wait a moment for animations
            time.sleep(1)

            # Take Screenshot
            screenshot_path = "verification/verification.png"
            page.screenshot(path=screenshot_path)
            print(f"Screenshot saved to {screenshot_path}")

        except Exception as e:
            print(f"Error during verification: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_command_palette()
