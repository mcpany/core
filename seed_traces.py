import asyncio
from playwright.async_api import async_playwright
import time

async def main():
    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        page = await browser.new_page()
        await page.goto("http://localhost:3000")

        # Trigger general seed
        res_seed = await page.evaluate('''async () => {
            const seedRequest = {
                upstream_services: [
                    {
                        id: "svc_01",
                        name: "Payment Gateway",
                        version: "v1.2.0",
                        http_service: {
                            address: "https://stripe.com"
                        }
                    }
                ]
            };
            const r = await fetch("/api/v1/debug/seed", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(seedRequest)
            });
            if (!r.ok) { return await r.text(); }
            return await r.json();
        }''')
        print("General Seeded:", res_seed)

        # Trigger debug traces seeding via fetch
        res = await page.evaluate('''async () => {
            const r = await fetch("/api/v1/debug/traces", { method: "POST" });
            if (!r.ok) { return await r.text(); }
            return await r.json();
        }''')
        print("Traces Seeded:", res)

        await page.reload()
        await asyncio.sleep(2)

        await page.screenshot(path="/home/jules/verification/screenshots/verification4.png", full_page=True)

        await browser.close()

if __name__ == "__main__":
    asyncio.run(main())
