const fs = require('fs');
const file = 'ui/src/lib/client.ts';
let code = fs.readFileSync(file, 'utf8');
code = code.replace(/@proto\/api\/v1\/registration/g, "../../../proto/api/v1/registration");
fs.writeFileSync(file, code);
