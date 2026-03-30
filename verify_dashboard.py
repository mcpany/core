from playwright.sync_api import sync_playwright
import urllib.request
import time

def run_cuj(page):
    try:
        req = urllib.request.Request("http://localhost:50050/api/v1/debug/traces", method="POST")
        with urllib.request.urlopen(req) as res:
            print("Seeded traces status:", res.status)
    except Exception as e:
        print(f"Could not seed via API: {e}")

    page.goto("http://localhost:9004/")
    page.wait_for_timeout(2000)

    # Try to add Recent Activity if not present
    try:
        page.locator('button:has-text("Add Widget")').click()
        page.wait_for_timeout(1000)
        locs = page.locator('button').all()
        for loc in locs:
            if "Recent Activity" in loc.inner_text():
                loc.click()
                break
        page.wait_for_timeout(1000)
        # Close sheet
        page.keyboard.press("Escape")
        page.wait_for_timeout(1000)
    except Exception as e:
        pass

    page.wait_for_timeout(3000)

    # Check if we are on Dashboard
    page.screenshot(path="/app/dashboard_after_seed.png")
    page.wait_for_timeout(1000)

    try:
        # Click on code refactor one specifically (should have the diff)
        locs = page.locator(".cursor-pointer").all()
        found = False
        for loc in locs:
            if "code-refactor" in loc.inner_text() or "github-pr-review" in loc.inner_text():
                loc.click()
                found = True
                page.wait_for_timeout(2000)
                # Scroll down a bit to see the diff block
                page.mouse.wheel(0, 500)
                page.wait_for_timeout(1000)
                page.screenshot(path="/app/dashboard_diff.png", full_page=True)
                page.wait_for_timeout(1000)
                break

        if not found and len(locs) > 0:
            locs[0].click()
            page.wait_for_timeout(1000)
            page.screenshot(path="/app/dashboard_expanded.png")
            page.wait_for_timeout(1000)

    except Exception as e:
        print(f"Could not expand trace: {e}")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            record_video_dir="/app/videos",
            viewport={'width': 1280, 'height': 800}
        )
        page = context.new_page()
        try:
            run_cuj(page)
        finally:
            context.close()
            browser.close()
