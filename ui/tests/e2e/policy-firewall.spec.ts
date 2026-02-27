import { test, expect } from '@playwright/test';

test('Policy Firewall E2E', async ({ page }) => {
  // 1. Navigate to Upstream Services
  await page.goto('/upstream-services');

  // 2. Open the seeded service (or create one if not found)
  // We'll try to find "policy-test-service". If not found, we assume we might need to create it manually in test,
  // but for now let's assume we can pick the first available service or "mcp-verified-service" which we saw in logs.
  // Actually, we can just create a new service via UI to be safe and independent.
  await page.click('button:has-text("Add Service")');
  await page.click('text=Empty Configuration');

  // 3. Configure Basic Service
  const uniqueName = `policy-firewall-test-${Date.now()}`;
  await page.fill('input[placeholder="My Service"]', uniqueName);
  await page.click('text=Connection');
  await page.click('text=Command Line (Stdio)');
  await page.fill('input[placeholder="docker"]', 'echo "test"');

  // 4. Go to Policies Tab
  await page.click('text=Policies');

  // 5. Interact with CallPolicyEditor
  // Default Action should be Allow All
  await expect(page.getByText('Allow All')).toBeVisible();

  // Add a Rule
  await page.click('button:has-text("Add Rule")');

  // Configure Rule: Deny specific tool
  // We need to find the inputs within the rule row.
  // The structure is roughly: Select(Action) | Input(NameRegex) | Input(ArgRegex)
  // We can scope by the rule container.

  // Select "Deny"
  // The select trigger might be tricky to target specifically if there are multiple.
  // We can try to target the one inside the draggable item.
  // Let's assume it's the first Select in the list.
  await page.click('div[role="combobox"]:has-text("Allow") >> nth=0');
  await page.click('text=Deny'); // Select Option

  // Set Tool Regex
  await page.fill('input[placeholder=".*"] >> nth=0', 'dangerous_tool');

  // 6. Test with Simulator
  // Default (should be allowed)
  await page.fill('input[id="sim-tool"]', 'safe_tool');
  await page.click('button:has-text("Test Rules")');
  await expect(page.getByText('Result:')).toBeVisible();
  await expect(page.getByText('Allowed')).toBeVisible(); // Default action

  // Blocked (should be denied)
  await page.fill('input[id="sim-tool"]', 'dangerous_tool');
  await page.click('button:has-text("Test Rules")');
  await expect(page.getByText('Denied')).toBeVisible();
  await expect(page.getByText('(Matched Rule #1)')).toBeVisible();

  // 7. Save Service
  await page.click('button:has-text("Save Changes")');
  await expect(page.getByText('Service Created')).toBeVisible();

  // 8. Verify Persistence (Reload and check)
  await page.reload();
  await page.click('text=Policies');
  await expect(page.getByDisplayValue('dangerous_tool')).toBeVisible();
});
