const fs = require('fs');
let file = 'server/tests/integration/examples/Dockerfile.timeserver';
let code = fs.readFileSync(file, 'utf8');
code = code.replace('WORKDIR /srv', 'WORKDIR /app/srv');
fs.writeFileSync(file, code);
