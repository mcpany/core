const fs = require('fs');
const filepath = 'ui/tests/e2e/audit-log.spec.ts';
let content = fs.readFileSync(filepath, 'utf8');

content = content.replace("expect(seedRes.ok()).toBeTruthy();", `
    if (!seedRes.ok()) {
      const text = await seedRes.text();
      console.log(\`DEBUG SEED FAILED: \${seedRes.status()} \${text}\`);
    }
    expect(seedRes.ok()).toBeTruthy();
`);
fs.writeFileSync(filepath, content);
