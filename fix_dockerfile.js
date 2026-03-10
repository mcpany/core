/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

const fs = require('fs');
let file = 'server/tests/integration/examples/Dockerfile.timeserver';
let code = fs.readFileSync(file, 'utf8');
code = code.replace('WORKDIR /srv', 'WORKDIR /app/srv');
fs.writeFileSync(file, code);
