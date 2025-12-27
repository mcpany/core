from playwright.sync_api import sync_playwright

def verify_metrics_dashboard():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        try:
            # Navigate to dashboard
            page.goto("http://localhost:3000")

            # Wait for metrics to load
            page.wait_for_selector("text=Total Requests")

            # Take screenshot of the dashboard with metrics
            page.screenshot(path="/home/jules/verification/dashboard_metrics.png")
            print("Screenshot saved to /home/jules/verification/dashboard_metrics.png")

        except Exception as e:
            print(f"Error: {e}")
            # Take screenshot of error state if possible
            try:
                page.screenshot(path="/home/jules/verification/error_state.png")
            except:
                pass
        finally:
            browser.close()

if __name__ == "__main__":
    verify_metrics_dashboard()
