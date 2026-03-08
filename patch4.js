const fs = require('fs');
const filepath = './server/docker/Dockerfile.server';
let content = fs.readFileSync(filepath, 'utf8');

// Undo WORKDIR /srv changes to isolate the issue
content = content.replace(/RUN mkdir -p \/srv\nWORKDIR \/srv\/server/g, 'WORKDIR /app/server');
content = content.replace(/RUN mkdir -p \/srv\nWORKDIR \/srv/g, 'WORKDIR /app');
content = content.replace(/\/srv/g, '/app');

// The original failure was `WORKDIR /srv` in `Dockerfile.timeserver` which I also modified. But the CI log says:
// ERROR: failed to build: failed to solve: process "/bin/sh -c CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH}     go build -o /server cmd/server/main.go" did not complete successfully: exit code: 2
fs.writeFileSync(filepath, content);
