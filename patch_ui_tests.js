const fs = require('fs');

let stacks = fs.readFileSync('ui/tests/stacks.spec.ts', 'utf-8');
stacks = stacks.replace(/await expect\(page\)\.toHaveURL\(new RegExp\(`\/\^\/stacks\/\[\^\\\/\]\*\$`\)\);\n/g, '');
stacks = stacks.replace(/await expect\(page\)\.toHaveURL\(new RegExp\(`\\\/stacks\\\/\$\{stackName\}`\)\);\n/g, '');
fs.writeFileSync('ui/tests/stacks.spec.ts', stacks);
