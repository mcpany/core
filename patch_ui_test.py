import re

with open("ui/tests/audit-logs.spec.ts", "r") as f:
    content = f.read()

# remove the mock completely
content = re.sub(r"    // Click it \(which triggers an export on backend\)\n    await page\.route\('\*\*/api/v1/audit/export\*', async route => \{\n        await route\.fulfill\(\{ status: 200, body: 'a,b,c\\n1,2,3' \}\);\n    \}\);\n\n", "", content)

# ensure wait for text works
content = content.replace("await expect(page.locator('text=echo_tool').first()).toBeVisible();", "await expect(page.locator('text=echo_tool').first()).toBeVisible({ timeout: 15000 });")

with open("ui/tests/audit-logs.spec.ts", "w") as f:
    f.write(content)
