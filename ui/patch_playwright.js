const fs = require('fs');
const filePath = 'playwright.config.ts';
let content = fs.readFileSync(filePath, 'utf8');

content = content.replace(
  "import { defineConfig, devices } from '@playwright/test';",
  "import { defineConfig, devices } from '@playwright/test';\nimport path from 'path';\n\nrequire('tsconfig-paths').register({\n  baseUrl: './',\n  paths: {\n    '@proto/*': ['./proto/*', '../proto/*'],\n    '../../../proto/*': ['../proto/*'],\n    '../proto/*': ['../proto/*'],\n    '../../proto/*': ['../proto/*'],\n    '../../../../proto/*': ['../proto/*'],\n    '../../../google/*': ['./node_modules/ts-proto-descriptors/dist/google/*']\n  }\n});"
);

fs.writeFileSync(filePath, content);
