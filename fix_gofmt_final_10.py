import re
import os
import glob

# Remove mocks in UI:
def ensure_seed(filepath):
    with open(filepath, "r") as f:
        content = f.read()

    # Add import { seedGlobalState } from './test-data'; if missing
    if "test-data" not in content and "seedGlobalState" not in content:
        content = content.replace("import { test, expect } from '@playwright/test';", "import { test, expect } from '@playwright/test';\nimport { seedGlobalState } from './test-data';")

        # Add seedGlobalState(request) in beforeEach
        if "test.beforeEach(async ({ page }) => {" in content:
            content = content.replace("test.beforeEach(async ({ page }) => {", "test.beforeEach(async ({ page, request }) => {\n    await seedGlobalState(request);")
        elif "test.beforeEach(async ({ page, request }) => {" in content:
            content = content.replace("test.beforeEach(async ({ page, request }) => {", "test.beforeEach(async ({ page, request }) => {\n    await seedGlobalState(request);")

        # Update login to use seeded user
        content = content.replace("await page.fill('input[name=\"username\"]', 'admin');", "await page.fill('input[name=\"username\"]', 'e2e-admin-core');")

    with open(filepath, "w") as f:
        f.write(content)

for root, _, files in os.walk("ui/tests/e2e"):
    for file in files:
        if file.endswith(".spec.ts"):
            ensure_seed(os.path.join(root, file))

def clean_file(path):
    with open(path, 'r') as f:
        content = f.read()

    new_content = ""
    i = 0
    while i < len(content):
        if content[i:i+17] == "await page.route(":
            paren_count = 1
            i += 17
            while i < len(content) and paren_count > 0:
                if content[i] == '(':
                    paren_count += 1
                elif content[i] == ')':
                    paren_count -= 1
                i += 1
            while i < len(content) and content[i] in [' ', '\t', '\n', ';']:
                i += 1
            continue
        else:
            new_content += content[i]
            i += 1

    with open(path, 'w') as f:
        f.write(new_content)

for path in glob.glob('ui/tests/e2e/**/*.spec.ts', recursive=True):
    clean_file(path)


# Traces UI mock specific
with open("ui/tests/e2e/traces.spec.ts", "r") as f:
    content = f.read()
    content = re.sub(r'// Mock Traces API for all tests in this suite\..*?}\);\n    }\);\n', "const res = await request.post('/api/v1/debug/traces', { data: {} });\n    expect(res.ok()).toBeTruthy();\n", content, flags=re.DOTALL)
with open("ui/tests/e2e/traces.spec.ts", "w") as f:
    f.write(content)

with open("ui/tests/e2e/live-trace.spec.ts", "r") as f:
    content = f.read()
    content = re.sub(r'// Mock traces API.*?}\);\n  }\);\n', "const res = await request.post('/api/v1/debug/traces', { data: {} });\n  expect(res.ok()).toBeTruthy();\n", content, flags=re.DOTALL)
    content = content.replace("test('Live Trace Inspector and Replay Flow', async ({ page }) => {", "test('Live Trace Inspector and Replay Flow', async ({ page, request }) => {")
with open("ui/tests/e2e/live-trace.spec.ts", "w") as f:
    f.write(content)

with open("ui/tests/e2e/settings.spec.ts", "r") as f:
    content = f.read()
    content = re.sub(r'// Mock Global Settings API.*?await page\.goto\(\'/settings\'\);', r'''await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { timeout: 30000 });
    await page.goto('/settings');''', content, flags=re.DOTALL)
with open("ui/tests/e2e/settings.spec.ts", "w") as f:
    f.write(content)

# Fix test data types
with open("ui/tests/e2e/test-data.ts", "r") as f:
    content = f.read()
    content = content.replace("import { ServiceTemplate } from '../../../proto/config/v1/service_template';", "// import { ServiceTemplate } from '../../../proto/config/v1/service_template';")
    content = content.replace("import { UpstreamServiceConfig } from '../../../proto/config/v1/upstream_service';", "// import { UpstreamServiceConfig } from '../../../proto/config/v1/upstream_service';")
    content = content.replace("import { User } from '../../../proto/config/v1/user';", "// import { User } from '../../../proto/config/v1/user';")
    content = content.replace("].map((service) => UpstreamServiceConfig.toJSON(UpstreamServiceConfig.fromJSON(service)));", "] as any;")
    content = content.replace("].map((template) => ServiceTemplate.toJSON(ServiceTemplate.fromJSON(template)));", "] as any;")
    content = content.replace("].map((user) => User.toJSON(User.fromJSON(user)));", "] as any;")
with open("ui/tests/e2e/test-data.ts", "w") as f:
    f.write(content)

# Fix backend skips
server_files = [
    "server/tests/public_api/the_meal_db_test.go",
    "server/tests/public_api/the_cocktail_db_test.go",
    "server/tests/public_api/genderize_test.go",
    "server/tests/public_api/agify_test.go",
    "server/tests/public_api/dog_facts_test.go",
    "server/tests/public_api/cat_facts_test.go",
    "server/tests/public_api/nationalize_test.go"
]

for file in server_files:
    with open(file, "r") as f:
        content = f.read()

    # Use exact replaces to avoid breaking gofmt structure
    content = content.replace('// \tt.Skip("Skipping test, no drinks found in response")', '')
    content = content.replace('\t\t// t.Skip("Skipping test, no drinks found in response")\n', '')
    content = content.replace('\t\t// t.Skip("Skipping test, no meals found in response")\n', '')
    content = content.replace('\t// t.SkipNow()\n', '')
    content = content.replace('\t// t.Skip("Skipping flaky cat facts test due to rate limiting issues")\n', '')
    content = content.replace('\t// t.SkipNow() // Removed skip\n', '')
    content = content.replace('\t\tt.Skip("Skipping test, no drinks found in response")\n', '')
    content = content.replace('\t\tt.Skip("Skipping test, no meals found in response")\n', '')
    content = content.replace('\tt.SkipNow()\n', '')
    content = content.replace('\tt.Skip("Skipping flaky cat facts test due to rate limiting issues")\n', '')


    # Check for empty ifs left by skip removal
    content = re.sub(r'if _, ok := .*?\[".*?"\]\.\(string\); ok \{\s*\}', '', content)
    content = re.sub(r'if .*?\[".*?"\] == nil \{\s*\}', '', content)

    # Check `TestUpstreamService_Genderize(t *testing.T) { //` -> `TestUpstreamService_Genderize(t *testing.T) {`
    content = re.sub(r'\(t \*testing\.T\)\s*{\s*//\s*\n\s*t\.SkipNow\(\)', r'(t *testing.T) {\n', content)
    content = re.sub(r'\(t \*testing\.T\)\s*{\s*//\s*\n', r'(t *testing.T) {\n', content)

    # Clean up util imports
    if "pkg/util" in content and "util." not in content:
        content = content.replace('\t"github.com/mcpany/core/server/pkg/util"\n', '')

    with open(file, "w") as f:
        f.write(content)

# Fix un-skipping comments
with open("ui/tests/e2e.spec.ts", "r") as f:
    content = f.read()
    content = content.replace("// We skip checking error details as it depends on runtime health check timing", "await expect(userService.locator('text=Healthy').or(userService.locator('text=Unhealthy'))).toBeVisible({ timeout: 15000 });")
with open("ui/tests/e2e.spec.ts", "w") as f:
    f.write(content)
