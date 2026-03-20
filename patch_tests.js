const fs = require('fs');
const filepath = 'ui/tests/e2e/test-data.ts';
let content = fs.readFileSync(filepath, 'utf8');

// The e2e tests are importing non-existent generated proto files,
// because they're not built. We'll replace the explicit imports with an 'any' cast for test payload.
content = content.replace("import { ServiceTemplate } from '../../../proto/config/v1/service_template';", "");
content = content.replace("import { UpstreamServiceConfig } from '../../../proto/config/v1/upstream_service';", "");
content = content.replace("import { User } from '../../../proto/config/v1/user';", "");

content = content.replace("ServiceTemplate[]", "any[]");
content = content.replace("UpstreamServiceConfig[]", "any[]");
content = content.replace("User[]", "any[]");

fs.writeFileSync(filepath, content);
