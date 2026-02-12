from playwright.sync_api import sync_playwright

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    page = browser.new_page()

    # Mock API responses
    page.route("**/api/v1/tools", lambda route: route.fulfill(
        status=200,
        content_type="application/json",
        body='{"tools": ' + str([
            {
                "name": f"tool_{i}",
                "description": f"Description for tool {i}",
                "serviceId": "service_1" if i % 2 == 0 else "service_2",
                "inputSchema": {},
                "outputSchema": {}
            } for i in range(500)
        ]).replace("'", '"') + '}'
    ))

    page.route("**/api/v1/services", lambda route: route.fulfill(
        status=200,
        content_type="application/json",
        body='{"services": [{"id": "service_1", "name": "Service 1"}, {"id": "service_2", "name": "Service 2"}]}'
    ))

    page.route("**/api/v1/dashboard/tool-usage", lambda route: route.fulfill(
        status=200,
        content_type="application/json",
        body='[]'
    ))

    # Also handle the initial dashboard page redirect or direct navigation
    # Navigating to /tools directly
    try:
        page.goto("http://localhost:9002/tools", timeout=60000)

        # Wait for table to load
        # Look for the virtualized table container or rows
        # Virtuoso creates a scroller div. We can wait for "tool_0" text.
        page.wait_for_selector("text=tool_0", timeout=30000)

        # Take a screenshot
        page.screenshot(path="verification/tools_page.png", full_page=True)
        print("Screenshot taken: verification/tools_page.png")

    except Exception as e:
        print(f"Error: {e}")
        page.screenshot(path="verification/error.png")
    finally:
        browser.close()

with sync_playwright() as playwright:
    run(playwright)
