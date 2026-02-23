from playwright.sync_api import sync_playwright, expect
import time

def test_agent_flow():
    with sync_playwright() as p:
        print("Launching browser...")
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Capture console logs
        page.on("console", lambda msg: print(f"Browser Console: {msg.text}"))
        page.on("pageerror", lambda err: print(f"Browser Error: {err}"))

        try:
            print("Navigating to Visualizer...")
            page.goto("http://localhost:9002/visualizer")

            # Wait for main container
            print("Waiting for page load...")
            page.wait_for_selector(".react-flow", timeout=10000)

            # Check for title
            expect(page.get_by_role("heading", name="Agent Flow Visualizer")).to_be_visible()

            # Find "Simulate Load" button
            seed_btn = page.get_by_role("button", name="Simulate Load")
            expect(seed_btn).to_be_visible()

            print("Clicking Simulate Load...")
            seed_btn.click()

            # Wait for backend to process and frontend to poll (it polls every 1s)
            print("Waiting for traffic animation...")
            page.wait_for_timeout(3000)

            # Screenshot
            print("Taking screenshot...")
            page.screenshot(path="verification/agent_flow_traffic_debug.png")
            print("Screenshot saved to verification/agent_flow_traffic_debug.png")

        except Exception as e:
            print(f"Test failed: {e}")
            page.screenshot(path="verification/failure.png")
            raise e
        finally:
            browser.close()

if __name__ == "__main__":
    test_agent_flow()
