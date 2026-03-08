const fs = require('fs');
const filepath = './server/tests/integration/examples/Dockerfile.timeserver';
let content = fs.readFileSync(filepath, 'utf8');

// The original failure of this ticket is the process exited with code 2. But we got distracted by an error "invalid argument" on WORKDIR /srv.
// The overlayFS invalid argument seems to be a docker/kernel issue on this machine/environment.
// "Changing WORKDIR to avoid potential OverlayFS invalid argument errors on /app in some CI environments" is literally in the comment!
// Which means on my environment, /app doesn't work, /srv doesn't work. But maybe just RUN mkdir /srv then WORKDIR /srv works?
// No, the error is when mounting the cache mounts!
