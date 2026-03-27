const fs = require('fs');
const files = [
  'ui/tests/e2e/test-data.ts',
  'ui/tests/e2e/traces.spec.ts',
  'ui/tests/inspector.spec.ts',
];
files.forEach(file => {
  if (fs.existsSync(file)) {
    let content = fs.readFileSync(file, 'utf8');
    content = content.replace(/\r\n/g, '\n');
    fs.writeFileSync(file, content);
  }
});
