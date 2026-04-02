const fs = require('fs');

const path = 'ui/tests/audit-logs.spec.ts';
let content = fs.readFileSync(path, 'utf8');

// Use the old mock since backend isn't ready
const oldMockBlock = `    await page.route('**/api/v1/audit/logs*', async route => {
        await route.fulfill({
            json: {
                entries: [
                    {
                        timestamp: new Date().toISOString(),
                        toolName: "get_users",
                        userId: "e2e-admin-core",
                        arguments: JSON.stringify({ "status": "all" }),
                        result: JSON.stringify([{"id": "1", "name": "Alice"}]),
                        duration: "230ms",
                        error: ""
                    }
                ]
            }
        });
    });`;

const newMockBlock = `    await page.route('**/api/v1/audit/logs*', async route => {
        await route.fulfill({
            json: {
                entries: [
                    {
                        timestamp: new Date().toISOString(),
                        toolName: "echo_tool",
                        userId: "e2e-admin-core",
                        arguments: JSON.stringify({ "hello": "world" }),
                        result: JSON.stringify({ "output": "world" }),
                        duration: "10ms",
                        error: ""
                    }
                ]
            }
        });
    });`;

content = content.replace(oldMockBlock, newMockBlock);

const expect1 = `await expect(page.locator('text=get_users').first()).toBeVisible();`;
const newExpect1 = `await expect(page.locator('text=echo_tool').first()).toBeVisible();`;
content = content.replace(expect1, newExpect1);

const expect2 = `await expect(page.locator('text=Alice').first()).toBeVisible();`;
const newExpect2 = `await expect(page.locator('text=world').first()).toBeVisible();`;
content = content.replace(expect2, newExpect2);


fs.writeFileSync(path, content, 'utf8');
console.log('Patched tests successfully');
