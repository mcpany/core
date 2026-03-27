/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

const fs = require('fs');
const path = 'ui/tests/e2e/traces.spec.ts';
let content = fs.readFileSync(path, 'utf8');
content = content.replace(/\s+$/g, '\n');
fs.writeFileSync(path, content);
