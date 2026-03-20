const fs = require('fs');
let testContent = fs.readFileSync('ui/tests/e2e/audit-log.spec.ts', 'utf8');
testContent = testContent.replace("// Rendered data has markdown text", "");
fs.writeFileSync('ui/tests/e2e/audit-log.spec.ts', testContent);
