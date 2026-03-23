const fs = require('fs');
const file = 'ui/tests/inspector.spec.ts';
let code = fs.readFileSync(file, 'utf8');
code = code.replace(/let wsSend: \(\(data: string\) => void\) \| null = null;/g, "let wsSend: any = null;");
fs.writeFileSync(file, code);
