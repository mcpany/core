const fs = require('fs');
let file = fs.readFileSync('ui/tests/e2e.spec.ts', 'utf-8');

file = file.replace(
    "import { seedGlobalState, seedTraffic, seedWebhooks } from './e2e/test-data';",
    "import { seedGlobalState, seedTraffic, seedWebhooks, seedHealth } from './e2e/test-data';"
);

file = file.replace(
    'await seedWebhooks(request);',
    'await seedWebhooks(request);\n    await seedHealth(request);'
);

fs.writeFileSync('ui/tests/e2e.spec.ts', file);
