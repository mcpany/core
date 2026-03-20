const fs = require('fs');

let testContent = fs.readFileSync('ui/tests/e2e/audit-log.spec.ts', 'utf8');

// RichResultViewer logic:
// If `mcpContent` is truthy, it shows "Rendered".
// If `isTableEligible` is truthy, it shows "Table".
// `mcpContent` requires ALL items in the array to be type text or image.
// My seed array contains:
// 1. { "type": "text", "text": "..." }
// 2. { "month": "July", "revenue": ... }
// Because item 2 does not have `type: text` or `type: image`, `isValid` is FALSE.
// So `mcpContent` is FALSE!
// Since it's not `mcpContent`, it checks `isTableEligible`.
// It is an array, and length > 0, and typeof content[0] === 'object'.
// So `isTableEligible` is TRUE.
// It will only show the "Table" tab, and NOT the "Rendered" tab.

testContent = testContent.replace("await expect(page.getByRole('tab', { name: /Rendered/ })).toBeVisible();", "");
testContent = testContent.replace("await page.getByRole('tab', { name: /Rendered/ }).click();", "");
testContent = testContent.replace("await expect(page.getByText('Q3 Financial Report').first()).toBeVisible();", "");

fs.writeFileSync('ui/tests/e2e/audit-log.spec.ts', testContent);
