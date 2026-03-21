import re
import glob
import os

for filepath in glob.glob('ui/tests/e2e/test-data.ts'):
    with open(filepath, 'r') as f:
        content = f.read()

    content = content.replace("import { ServiceTemplate } from '../../../proto/config/v1/service_template';", "import type { ServiceTemplate } from '../../src/lib/client';")
    content = content.replace("import { UpstreamServiceConfig } from '../../../proto/config/v1/upstream_service';", "import type { UpstreamServiceConfig } from '../../src/lib/client';")
    content = content.replace("import { User } from '../../../proto/config/v1/user';", "import type { User } from '../../src/components/users/user-list';")

    with open(filepath, 'w') as f:
        f.write(content)
