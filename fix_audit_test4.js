const fs = require('fs');
let testContent = fs.readFileSync('ui/tests/e2e/audit-log.spec.ts', 'utf8');

testContent = testContent.replace(`    if (!seedRes.ok()) {
      const text = await seedRes.text();
      console.log(\`DEBUG SEED FAILED: \${seedRes.status()} \${text}\`);
    }
    expect(seedRes.ok()).toBeTruthy();`, `    expect(seedRes.ok()).toBeTruthy();`);

fs.writeFileSync('ui/tests/e2e/audit-log.spec.ts', testContent);
