from playwright.sync_api import sync_playwright

def verify_network_graph():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Navigate to the network page
        page.goto("http://localhost:9002/network")

        # Wait for the graph to load (look for a known element)
        page.wait_for_selector(".react-flow__renderer")

        # Take a screenshot
        screenshot_path = ".audit/ui/2025-12-25/network_graph.png"
        page.screenshot(path=screenshot_path)
        print(f"Screenshot saved to {screenshot_path}")

        browser.close()

if __name__ == "__main__":
    verify_network_graph()
