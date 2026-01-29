from playwright.sync_api import sync_playwright
import time

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        try:
            page.goto("http://localhost:9002/playground")

            # Execute tool
            page.get_by_placeholder("Enter command or select a tool...").fill("user-service.list_users {}")
            page.keyboard.press("Enter")

            # Wait for Table
            page.get_by_role("button", name="Table").wait_for(timeout=20000)

            # Screenshot
            page.screenshot(path="verification_table.png")
            print("Screenshot taken.")
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification_error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
