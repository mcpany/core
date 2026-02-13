from playwright.sync_api import sync_playwright
import time

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            print("Navigating to dashboard...")
            page.goto("http://localhost:9112")

            # Wait for content
            print("Waiting for 'Welcome to MCP Any!'...")
            page.wait_for_selector("text=Welcome to MCP Any!", timeout=30000)

            # Take screenshot
            path = "verification/onboarding_step1.png"
            page.screenshot(path=path)
            print(f"Screenshot saved to {path}")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
