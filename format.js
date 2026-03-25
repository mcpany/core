const fs = require('fs');
let file = fs.readFileSync('ui/src/components/register-service-dialog.test.tsx', 'utf8');
file = file.replace(/\n\s*\n\s*\n/g, '\n\n');
file = file.replace(/\n+$/g, '\n');
fs.writeFileSync('ui/src/components/register-service-dialog.test.tsx', file);
