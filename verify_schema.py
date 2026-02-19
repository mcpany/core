
from playwright.sync_api import sync_playwright

def verify_schema_playground():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            # Assuming dev server is running on port 9002 (as per package.json)
            page.goto("http://localhost:9002/playground/schema")

            # Check for title
            page.wait_for_selector("h2:has-text('Schema Validator')")

            # Check for Editor
            page.wait_for_selector("text=Editor")

            # Check for Monaco Editor (class 'monaco-editor')
            # It might take a moment to load
            page.wait_for_selector(".monaco-editor", timeout=10000)

            print("Schema Playground verified successfully.")
        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path="schema_verification_failure.png")
            raise e
        finally:
            browser.close()

if __name__ == "__main__":
    verify_schema_playground()
