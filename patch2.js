const fs = require('fs');
const filepath = './server/tests/integration/examples/Dockerfile.timeserver';
let content = fs.readFileSync(filepath, 'utf8');

content = content.replace(/WORKDIR \/srv/g, 'WORKDIR /srv\nRUN mkdir -p /srv');
fs.writeFileSync(filepath, content);
