from playwright.sync_api import sync_playwright

def verify(page):
    print("Navigating to dashboard...")
    page.goto("http://localhost:3000/")

    print("Waiting for network graph...")
    try:
        page.wait_for_selector(".react-flow", state="visible", timeout=60000)
        print("Network graph found.")
    except Exception as e:
        print(f"Wait failed: {e}")

    print("Taking screenshot...")
    # Scroll to bottom to ensure everything renders?
    # React Flow might use virtualization or canvas, but it should be fine.
    page.screenshot(path="verification.png", full_page=True)
    print("Screenshot saved to verification.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        # Set a large viewport
        page.set_viewport_size({"width": 1280, "height": 2000})
        try:
            verify(page)
        except Exception as e:
            print(f"Error: {e}")
        finally:
            browser.close()
