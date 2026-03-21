const fs = require('fs');
let content = fs.readFileSync('ui/src/mocks/proto/mock-proto.ts', 'utf8');
content = content.replace('export const CallPolicyRule = {};', '/**\n * Mock type placeholders for policy-related proto messages.\n */\nexport const CallPolicyRule = {};');
content = content.replace('export const ExportPolicy = {};', '/**\n * Mock type placeholders for policy-related proto messages.\n */\nexport const ExportPolicy = {};');
content = content.replace('export const ExportRule = {};', '/**\n * Mock type placeholders for policy-related proto messages.\n */\nexport const ExportRule = {};');
fs.writeFileSync('ui/src/mocks/proto/mock-proto.ts', content);
