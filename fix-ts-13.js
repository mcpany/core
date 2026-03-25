const fs = require('fs');

function replaceFile(path, search, replace) {
  let content = fs.readFileSync(path, 'utf8');
  content = content.replace(search, replace);
  fs.writeFileSync(path, content);
}

replaceFile('ui/src/lib/config-utils.test.ts', "const secretHandlingMode: any =", "const secretHandlingMode: unknown =");
replaceFile('ui/src/lib/config-utils.test.ts', "const secretMap: any =", "const secretMap: unknown =");
replaceFile('ui/src/lib/config-utils.ts', "export function redactSecrets(obj: any): any {", "export function redactSecrets(obj: unknown): unknown {");

replaceFile('ui/src/lib/mcp-unwrap.ts', "export function unwrapResult(response: any): any {", "export function unwrapResult(response: unknown): unknown {");
