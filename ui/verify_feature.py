
from playwright.sync_api import sync_playwright, expect
import time

def verify_prompt_workbench():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={"width": 1280, "height": 800})
        page = context.new_page()

        try:
            print("Navigating to Prompts Page (Port 9002)...")
            # Wait for next dev server to be ready
            for i in range(15):
                try:
                    page.goto("http://localhost:9002/prompts", timeout=5000)
                    break
                except:
                    print(f"Waiting for server... {i+1}")
                    time.sleep(3)

            print("Waiting for prompt list...")
            expect(page.get_by_text("Prompt Library")).to_be_visible(timeout=30000)

            page.screenshot(path="/home/jules/verification/prompt_workbench_initial.png")
            print("Initial screenshot taken.")

            # Check for empty state or prompts
            if page.get_by_text("No prompts found").is_visible():
                print("No prompts found (expected if no backend).")
            else:
                # Try to click the first button
                # Using a broad selector might be flaky, but acceptable for verification
                first_prompt = page.locator("div[class*='border-r'] button").first
                if first_prompt.is_visible():
                    first_prompt.click()
                    print("Clicked first prompt.")
                    time.sleep(1)

                    # Check for details
                    expect(page.get_by_text("Configuration")).to_be_visible()
                    print("Details view visible.")

            page.screenshot(path="/home/jules/verification/prompt_workbench_final.png")
            print(f"Final screenshot saved to /home/jules/verification/prompt_workbench_final.png")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="/home/jules/verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_prompt_workbench()
