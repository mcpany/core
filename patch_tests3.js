const fs = require('fs');
const filepath = 'ui/tests/e2e/test-data.ts';
let content = fs.readFileSync(filepath, 'utf8');

// Some of the other imports might be broken as well if there's any left.
// Actually, let's look at the actual error for seedGlobalState 500 status.
console.log("Checking why seedGlobalState threw 500");
