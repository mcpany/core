from playwright.sync_api import sync_playwright, expect

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        print("Navigating to Stacks page...")
        page.goto("http://localhost:9002/stacks")

        # Expect "No stacks defined" or the list
        page.screenshot(path="verification/stacks_list.png")
        print("Screenshot of list taken.")

        print("Clicking Add Stack...")
        page.get_by_role("link", name="Add Stack").click()

        expect(page.get_by_role("heading", name="New Stack")).to_be_visible()

        # Verify inputs exist
        page.fill("input[id='name']", "test-stack")
        page.fill("textarea[id='description']", "My Test Stack")

        # Editor interaction is tricky with Monaco, but we can verify it exists
        # Monaco usually puts a textarea with class "inputarea" but it's hidden.
        # We can just verify the visual container exists.

        page.screenshot(path="verification/stacks_new.png")
        print("Screenshot of new stack page taken.")

        browser.close()

if __name__ == "__main__":
    run()
