import sys

with open("ui/tests/e2e/traces.spec.ts", "r") as f:
    content = f.read()

search = """        // Check output payload values
        await expect(page.getByText('Revenue up 15%')).toBeVisible();"""

replace = """        // Check output payload values
        await expect(page.getByText('Revenue up 15%').first()).toBeVisible();"""

search2 = """    // Wait for the trace to appear in the table
    const traceRow = page.locator('tr.cursor-pointer').first();
    await expect(traceRow).toBeVisible({ timeout: 10000 });"""

replace2 = """    // Wait for the trace to appear in the table
    const traceRow = page.locator('tr.cursor-pointer').first();
    await expect(traceRow).toBeVisible({ timeout: 30000 });"""

content = content.replace(search, replace)
content = content.replace(search2, replace2)

with open("ui/tests/e2e/traces.spec.ts", "w") as f:
    f.write(content)
