const fs = require('fs');
const file = 'ui/src/tests/tools.spec.ts';
if (fs.existsSync(file)) {
  let code = fs.readFileSync(file, 'utf8');
  code = code.replace(/ws\.send\(data\)/g, "ws.send(data)");
  fs.writeFileSync(file, code);
}
