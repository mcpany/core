
import asyncio
from playwright.async_api import async_playwright, expect

async def verify_stats_page():
    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        page = await browser.new_page()

        # Go to the stats page
        await page.goto("http://localhost:9002/stats")

        # Wait for the dashboard to load
        await expect(page.get_by_text("Analytics & Stats")).to_be_visible()

        # Check for key elements
        await expect(page.get_by_text("Total Requests")).to_be_visible()
        await expect(page.get_by_text("Avg Latency")).to_be_visible()
        await expect(page.get_by_text("Error Rate")).to_be_visible()

        # Check tabs
        await expect(page.get_by_text("Overview")).to_be_visible()
        await expect(page.get_by_text("Performance")).to_be_visible()
        await expect(page.get_by_text("Errors")).to_be_visible()

        # Wait a bit for charts to animate (if any)
        await page.wait_for_timeout(2000)

        # Take a screenshot
        await page.screenshot(path=".audit/ui/2025-12-30/stats_analytics.png", full_page=True)

        await browser.close()

if __name__ == "__main__":
    asyncio.run(verify_stats_page())
