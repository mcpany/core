const fs = require('fs');
let code = fs.readFileSync('server/go.mod', 'utf-8');
code = code.replace(/github\.com\/mcpany\/core\/proto.*$/m, 'github.com/mcpany/core/proto v0.0.0');
fs.writeFileSync('server/go.mod', code);
