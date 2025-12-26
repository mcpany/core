
import asyncio
from playwright.async_api import async_playwright, expect

async def run():
    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        page = await browser.new_page()

        # Navigate to the network page
        # Note: The dev server port is 9002 based on package.json
        try:
            await page.goto("http://localhost:9002/network", timeout=60000)

            # Wait for the graph to load (look for the "Network Graph" title)
            await expect(page.get_by_text("Network Graph")).to_be_visible()

            # Wait a bit for the graph rendering animation
            await page.wait_for_timeout(2000)

            # Interact: Click on a node (e.g., "MCP Host")
            # ReactFlow nodes often don't have standard roles, so we might need text locators
            # Use exact=True to avoid matching the description text
            await page.get_by_text("MCP Host", exact=True).click()

            # Wait for the details panel to appear
            await expect(page.get_by_text("ID: host")).to_be_visible()

            # Take a screenshot
            await page.screenshot(path="verification/network_graph.png")
            print("Screenshot saved to verification/network_graph.png")

        except Exception as e:
            print(f"Error: {e}")
            # Take a screenshot anyway to debug
            await page.screenshot(path="verification/error.png")

        await browser.close()

if __name__ == "__main__":
    asyncio.run(run())
