from playwright.sync_api import Page, expect, sync_playwright
import time

def test_user_creation(page: Page):
    print("Navigating to dashboard...")
    page.goto("http://localhost:9002")

    # Wait for loading
    page.wait_for_timeout(5000)

    # Check if we are redirected to onboarding or dashboard
    if "onboarding" in page.url or "Connect" in page.content():
        print("Onboarding detected.")

    print("Navigating to Users page...")
    page.goto("http://localhost:9002/users")

    # Wait for page load
    page.wait_for_selector("h2:has-text('Users')", timeout=10000)

    print("Opening Add User sheet...")
    page.click("button:has-text('Add User')")

    # Wait for sheet
    page.wait_for_selector("div[role='dialog']", timeout=5000)

    print("Filling form...")
    page.fill("input[name='id']", "testuser_e2e")

    # Password tab should be selected by default
    page.fill("input[name='password']", "securepassword123")

    # Take screenshot of the form
    page.screenshot(path="/home/jules/verification/user_form.png")

    print("Submitting...")
    page.click("button:has-text('Save Changes')")

    # Wait for toast or list update
    page.wait_for_timeout(2000)

    # Verify user appears in list
    expect(page.get_by_text("testuser_e2e")).to_be_visible()

    # Take final screenshot
    page.screenshot(path="/home/jules/verification/user_list.png")
    print("Success!")

if __name__ == "__main__":
    with sync_playwright() as p:
        print("Launching browser...")
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            test_user_creation(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="/home/jules/verification/error.png")
        finally:
            browser.close()
