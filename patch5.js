const fs = require('fs');
const file = 'ui/tsconfig.json';
let code = fs.readFileSync(file, 'utf8');
code = code.replace(/"\.\.\/bazel-bin\/proto\/\*"/g, `"../bazel-bin/proto/*", "../proto/*"`);
fs.writeFileSync(file, code);
