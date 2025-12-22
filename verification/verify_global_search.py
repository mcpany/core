from playwright.sync_api import Page, expect, sync_playwright

def verify_global_search(page: Page):
    # 1. Navigate to the dashboard
    page.goto("http://localhost:3000")

    # 2. Open Global Search with Cmd+K (simulated)
    # Note: Cmd+K can be tricky to simulate reliably across OS, but we'll try standard way
    page.keyboard.press("Meta+k")

    # Wait for dialog to appear
    dialog = page.get_by_role("dialog")
    expect(dialog).to_be_visible()

    # 3. Search for "Services"
    search_input = page.get_by_placeholder("Type a command or search...")
    search_input.fill("Services")

    # 4. Verify "Services" option is visible and select it
    services_option = page.get_by_role("option", name="Services").first
    expect(services_option).to_be_visible()

    # Take screenshot of the open menu
    page.screenshot(path="verification/global_search_open.png")

    # Click it
    services_option.click()

    # 5. Verify navigation to /services
    expect(page).to_have_url("http://localhost:3000/services")

    # Screenshot of destination
    page.screenshot(path="verification/services_page.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_global_search(page)
        finally:
            browser.close()
