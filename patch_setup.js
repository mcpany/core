const fs = require('fs');
const path = 'ui/src/tests/setup.ts';
let code = fs.readFileSync(path, 'utf8');

code = code.replace(
  `json: () => Promise.resolve([]),`,
  `json: () => Promise.resolve([]),
     text: () => Promise.resolve("[]"),
     headers: new Headers(),`
);

fs.writeFileSync(path, code);
