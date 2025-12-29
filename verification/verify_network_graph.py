from playwright.sync_api import sync_playwright

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        page.goto("http://localhost:9002/network")

        # Wait for the graph to load (canvas element or similar)
        try:
            # Check for the topology card
            page.wait_for_selector("text=Network Topology", timeout=10000)
            print("Topology card found")

            # Check for some nodes
            page.wait_for_selector("text=MCP Any Core", timeout=5000)
            print("MCP Core node found")

            # Check for panel legend
            page.wait_for_selector("text=Agent (Active)", timeout=5000)
            print("Legend found")

            page.screenshot(path="verification/network_graph.png")
            print("Screenshot saved to verification/network_graph.png")
        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path="verification/error.png")

        browser.close()

if __name__ == "__main__":
    run()
