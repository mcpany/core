const fs = require('fs');

const file = 'ui/tests/e2e/user_management.spec.ts';
let content = fs.readFileSync(file, 'utf-8');

content = content.replace("const row = page.locator('tr').filter({ hasText: 'test-api-user' });", "const row = page.getByTestId('user-row-test-api-user');");

fs.writeFileSync(file, content);
