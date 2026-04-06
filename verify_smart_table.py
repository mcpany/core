from playwright.sync_api import sync_playwright

def run_cuj(page):
    page.goto("http://localhost:9003/tools")
    page.wait_for_timeout(3000)

    # Search for get_users
    page.get_by_placeholder("Search tools...").fill("get_users")
    page.wait_for_timeout(1000)

    # Click inspect
    page.get_by_role("button", name="Inspect").first.click()
    page.wait_for_timeout(1000)

    # Click execute
    dialog = page.get_by_role("dialog")
    dialog.get_by_role("button", name="Execute").click()
    page.wait_for_timeout(2000)

    # Make sure we are on Table tab
    table_tab = dialog.get_by_role("tab", name="Table")
    if table_tab.get_attribute("aria-selected") != "true":
        table_tab.click()
    page.wait_for_timeout(1000)

    # Click eye icon for object (should be [ ] Array(3) or similar but it's rendering inside)
    # The actual table cells will contain the Eye icon for objects if any, but get_users returns flat objects.
    # Wait, get_users returns an array of objects. The SmartTable renders the array of objects as rows.
    # Let's see if there is any long text.
    # get_users has a long_text column.

    # Click the expand button for the long text
    expand_btn = page.locator('button:has(.lucide-expand)').first
    if expand_btn.is_visible():
        expand_btn.click()
        page.wait_for_timeout(1000)

    # Take screenshot
    page.screenshot(path="/home/jules/verification/screenshots/smart_table.png")
    page.wait_for_timeout(1000)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            record_video_dir="/home/jules/verification/videos"
        )
        page = context.new_page()
        try:
            run_cuj(page)
        finally:
            context.close()
            browser.close()
