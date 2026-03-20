const fs = require('fs');

let content = fs.readFileSync('ui/tests/e2e/services.spec.ts', 'utf8');

// Add import correctly
if (!content.includes('import { seedGlobalState } from')) {
    content = "import { seedGlobalState } from './test-data';\n" + content;
}

fs.writeFileSync('ui/tests/e2e/services.spec.ts', content);
