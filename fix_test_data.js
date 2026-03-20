const fs = require('fs');

let content = fs.readFileSync('ui/tests/e2e/test-data.ts', 'utf8');

content = content.replace(/import \{ ServiceTemplate \} from '\.\.\/\.\.\/\.\.\/proto\/config\/v1\/service_template';/g, "// import { ServiceTemplate } from '../../../proto/config/v1/service_template';");
content = content.replace(/import \{ UpstreamServiceConfig \} from '\.\.\/\.\.\/\.\.\/proto\/config\/v1\/upstream_service';/g, "// import { UpstreamServiceConfig } from '../../../proto/config/v1/upstream_service';");
content = content.replace(/import \{ User \} from '\.\.\/\.\.\/\.\.\/proto\/config\/v1\/user';/g, "// import { User } from '../../../proto/config/v1/user';");

content = content.replace(/UpstreamServiceConfig\.toJSON\(UpstreamServiceConfig\.fromJSON\(service\)\)/g, "service");
content = content.replace(/ServiceTemplate\.toJSON\(ServiceTemplate\.fromJSON\(template\)\)/g, "template");
content = content.replace(/User\.toJSON\(User\.fromJSON\(user\)\)/g, "user");

fs.writeFileSync('ui/tests/e2e/test-data.ts', content);
