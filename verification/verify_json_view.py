from playwright.sync_api import sync_playwright

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()

    # Go to Playground (usually first tool is selected or default)
    # The URL is likely /playground
    print("Navigating to /playground...")
    try:
        page.goto("http://localhost:9002/playground", timeout=30000)
    except Exception as e:
        print(f"Failed to load /playground: {e}")
        # Try / if playground fails
        try:
             page.goto("http://localhost:9002/", timeout=30000)
             print("Navigated to / instead")
        except:
             print("Failed to navigate anywhere")
             browser.close()
             return

    # Wait for the page to load
    try:
        page.wait_for_load_state("networkidle", timeout=5000)
    except:
        print("Network idle timed out, continuing...")

    # Take a screenshot of the initial page
    page.screenshot(path="verification/playground_initial.png")
    print("Screenshot saved to verification/playground_initial.png")

    # Look for "Schema" tab in ToolForm
    # ToolForm has Tabs with "Schema" trigger.
    # It might be in a tablist.
    # Check if we are on playground first
    if "playground" not in page.url:
         # Navigate to playground manually if needed?
         print(f"Current URL: {page.url}")

    # Try to find Schema tab
    try:
        schema_tab = page.get_by_role("tab", name="Schema")
        if schema_tab.count() > 0:
            print("Clicking Schema tab...")
            schema_tab.click()

            # Wait for JsonView to render (Tree mode by default)
            # Look for "Tree" button which indicates JsonView toolbar is present
            tree_btn = page.get_by_role("button", name="Tree")
            try:
                tree_btn.wait_for(state="visible", timeout=5000)
                print("JsonView loaded in Tree mode.")
            except:
                print("Could not find JsonView toolbar (Tree button).")
                page.screenshot(path="verification/failed_to_find_tree.png")
                # Maybe try to find raw button directly?

            # Now click "Raw" button to trigger lazy loading of SyntaxHighlighter
            raw_btn = page.get_by_role("button", name="Raw")
            if raw_btn.count() > 0:
                print("Clicking Raw button...")
                raw_btn.click()

                # Wait for content to appear (or loading)
                page.wait_for_timeout(1000) # Give it a second to lazy load

                # Take a screenshot after clicking Raw
                page.screenshot(path="verification/json_view_raw.png")
                print("Screenshot saved to verification/json_view_raw.png")
            else:
                print("Raw button not found")

        else:
            print("Schema tab not found.")
            # Maybe look for any JsonView?
            # Or assume we are not on playground page.
    except Exception as e:
        print(f"Error interacting with page: {e}")
        page.screenshot(path="verification/error_state.png")

    browser.close()

with sync_playwright() as p:
    run(p)
