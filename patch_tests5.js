const fs = require('fs');
const filepath = 'ui/tests/e2e/test-data.ts';
let content = fs.readFileSync(filepath, 'utf8');

// I replaced the mapping logic earlier with `];` which means the definitions might be slightly misconfigured or missing now, because I removed `.map()`
// Wait, I replaced `].map(...)` with `];`. Let's check what services, templates, users are.
console.log(content.substring(content.indexOf('const services ='), content.indexOf('const templates =')));
