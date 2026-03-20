/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

const fs = require('fs');

const file = 'ui/src/components/users/user-list.tsx';
let content = fs.readFileSync(file, 'utf-8');


content = content.replace(/itemContent=\{\(index, user\) => \( <React.Fragment key=\{user\.id\}>/g, 'itemContent={(index, user) => (');
content = content.replace(/<\/React\.Fragment>\n                            \)/g, ')');


fs.writeFileSync(file, content);
