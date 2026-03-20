const fs = require('fs');
const filepath = 'ui/tests/e2e/test-data.ts';
let content = fs.readFileSync(filepath, 'utf8');

console.log("Restoring the original test-data.ts from git to ensure it works properly");
