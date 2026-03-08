const fs = require('fs');
const filepath = './server/docker/Dockerfile.server';
let content = fs.readFileSync(filepath, 'utf8');

// Change WORKDIR /app to WORKDIR /srv to see if it fixes it
content = content.replace(/WORKDIR \/app/g, 'RUN mkdir -p /srv\nWORKDIR /srv');
content = content.replace(/WORKDIR \/app\/server/g, 'WORKDIR /srv/server');
content = content.replace(/\/app/g, '/srv'); // fix paths like COPY /srv/server /server
fs.writeFileSync(filepath, content);
