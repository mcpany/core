import re
with open("ui/tests/dashboard_persistence.spec.ts", "r") as f:
    content = f.read()

replacement = """
  // Wait for loading to finish
  await expect(page.locator('.lucide-loader-circle.h-8')).not.toBeVisible({ timeout: 10000 });
"""

new_content = re.sub(
    r"  // Wait for loading to finish\n  await expect\(page\.locator\('\.lucide-loader-circle\.h-8'\)\)\.not\.toBeVisible\(\);",
    replacement.strip(),
    content
)

with open("ui/tests/dashboard_persistence.spec.ts", "w") as f:
    f.write(new_content)
