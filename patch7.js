const fs = require('fs');
const file = 'ui/tests/e2e/test-data.ts';
let code = fs.readFileSync(file, 'utf8');
code = code.replace(/\.\.\/\.\.\/\.\.\/proto\/config\/v1\/service_template/g, "@proto/config/v1/service_template");
code = code.replace(/\.\.\/\.\.\/\.\.\/proto\/config\/v1\/upstream_service/g, "@proto/config/v1/upstream_service");
code = code.replace(/\.\.\/\.\.\/\.\.\/proto\/config\/v1\/user/g, "@proto/config/v1/user");
fs.writeFileSync(file, code);

const file2 = 'ui/src/lib/client.ts';
let code2 = fs.readFileSync(file2, 'utf8');
code2 = code2.replace(/\.\.\/\.\.\/\.\.\/proto\/api\/v1\/registration/g, "@proto/api/v1/registration");
fs.writeFileSync(file2, code2);
