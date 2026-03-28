const fs = require('fs');
const file = 'ui/tests/tools.spec.ts';
let content = fs.readFileSync(file, 'utf8');

// There is one remaining `await page.waitForTimeout(1000)` in the test block
content = content.replace(/await page\.waitForTimeout\(1000\)/g, 'await new Promise(r => setTimeout(r, 1000))');

fs.writeFileSync(file, content);
