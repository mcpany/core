const fs = require('fs');
const filepath = 'ui/tests/e2e/test-data.ts';
let content = fs.readFileSync(filepath, 'utf8');

// Replace the mapping lines that use proto classes
content = content.replace(/\]\.map\(\(service\) => UpstreamServiceConfig\.toJSON\(UpstreamServiceConfig\.fromJSON\(service\)\)\);/g, "];");
content = content.replace(/\]\.map\(\(template\) => ServiceTemplate\.toJSON\(ServiceTemplate\.fromJSON\(template\)\)\);/g, "];");
content = content.replace(/\]\.map\(\(user\) => User\.toJSON\(User\.fromJSON\(user\)\)\);/g, "];");

fs.writeFileSync(filepath, content);
