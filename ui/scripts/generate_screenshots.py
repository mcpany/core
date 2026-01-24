from playwright.sync_api import sync_playwright, expect

def verify_users():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            # Go to Users page
            print("Navigating to users page...")
            page.goto("http://localhost:9002/users")

            # Wait for content
            print("Waiting for content...")
            page.wait_for_selector("text=Users", timeout=10000)

            # Click Add User
            print("Clicking Add User...")
            page.get_by_role("button", name="Add User").click()

            # Check for Role dropdown (it defaults to Viewer)
            print("Checking dropdown...")
            # Ideally we find the trigger. The trigger usually has role "combobox" in shadcn/radix
            combobox = page.get_by_role("combobox")
            expect(combobox).to_be_visible()
            expect(combobox).to_contain_text("Viewer")

            # Click the dropdown
            combobox.click()

            # Check options
            expect(page.get_by_role("option", name="Admin")).to_be_visible()
            expect(page.get_by_role("option", name="Editor")).to_be_visible()
            expect(page.get_by_role("option", name="Viewer")).to_be_visible()

            # Take screenshot of the dialog with dropdown open
            page.screenshot(path="ui/verification.png")
            print("Screenshot saved to ui/verification.png")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="ui/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_users()
