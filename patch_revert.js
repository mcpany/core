const fs = require('fs');
const filepath = './server/tests/integration/examples/Dockerfile.timeserver';
let content = fs.readFileSync(filepath, 'utf8');

// Undo WORKDIR /srv changes
content = content.replace(/WORKDIR \/srv\nRUN mkdir -p \/srv/g, 'WORKDIR /srv');

fs.writeFileSync(filepath, content);
