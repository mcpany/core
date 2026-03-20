const fs = require('fs');

let config = fs.readFileSync('ui/playwright.config.ts', 'utf8');

// Insert require('tsconfig-paths/register') at the top
if (!config.includes('tsconfig-paths/register')) {
    config = "require('tsconfig-paths/register');\n" + config;
    fs.writeFileSync('ui/playwright.config.ts', config);
}
