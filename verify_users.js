const { chromium } = require('playwright');

async function seedUser() {
    const url = "http://localhost:50050/api/v1/users";
    const headers = { "X-API-Key": "test-token", "Content-Type": "application/json" };
    const user = {
        id: "verify-admin",
        authentication: {
            basic_auth: {
                username: "verify-admin",
                password_hash: "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a"
            }
        },
        roles: ["admin"]
    };
    try {
        const res = await fetch(url, { method: 'POST', headers, body: JSON.stringify({ user }) });
        if (res.ok) console.log("Seeded user: verify-admin");
        else console.log("Failed to seed user:", res.status, await res.text());
    } catch (e) {
        console.log("Failed to seed user:", e);
    }
}

async function run() {
    await seedUser();
    const browser = await chromium.launch();
    const context = await browser.newContext();
    const page = await context.newPage();

    try {
        // 1. Login
        await page.goto("http://localhost:9002/login");
        await page.fill('input[name="username"]', 'verify-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await page.waitForURL("http://localhost:9002/", { timeout: 15000 });

        // 2. Go to Users
        await page.goto("http://localhost:9002/users");

        // 3. Open Add User Sheet
        await page.click('button:has-text("Add User")');

        // 4. Fill some data to show form
        await page.fill('input[name="id"]', 'new-user');

        // 5. Take screenshot
        await page.screenshot({ path: "/home/jules/verification/verification.png" });
        console.log("Screenshot saved to /home/jules/verification/verification.png");
    } catch (e) {
        console.error(e);
    } finally {
        await browser.close();
    }
}

run();
