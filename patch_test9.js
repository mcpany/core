const fs = require('fs');
const filepath = 'ui/tests/e2e/audit-log.spec.ts';
let content = fs.readFileSync(filepath, 'utf8');

// Update audit-log.spec.ts to just check if seedRes is ok, but if it is 500, log the text.
content = content.replace("expect(seedRes.ok()).toBeTruthy();", `
    if (!seedRes.ok()) {
      const text = await seedRes.text();
      console.log(\`DEBUG SEED FAILED: \${seedRes.status()} \${text}\`);
    }
    expect(seedRes.ok()).toBeTruthy();
`);
fs.writeFileSync(filepath, content);
