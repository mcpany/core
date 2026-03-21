import re
import glob
import os

for filepath in glob.glob('ui/tests/e2e/test-data.ts'):
    with open(filepath, 'r') as f:
        content = f.read()

    content = content.replace("import { ServiceTemplate } from '../../../proto/config/v1/service_template';", "import type { ServiceTemplate } from '../../src/lib/client';")
    content = content.replace("import { UpstreamServiceConfig } from '../../../proto/config/v1/upstream_service';", "import type { UpstreamServiceConfig } from '../../src/lib/client';")
    content = content.replace("import { User } from '../../../proto/config/v1/user';", "import type { User } from '../../src/components/users/user-list';")

    # We also need to fix the TS type references that are no longer valid if we import type
    content = content.replace("(svc as unknown as UpstreamServiceConfig)", "(svc as any)")
    content = content.replace("svc: UpstreamServiceConfig", "svc: any")
    content = content.replace("serviceConfig: UpstreamServiceConfig", "serviceConfig: any")
    content = content.replace("template: ServiceTemplate", "template: any")
    content = content.replace("user: User", "user: any")

    with open(filepath, 'w') as f:
        f.write(content)
