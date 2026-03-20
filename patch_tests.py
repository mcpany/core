import re
import glob
import os

for filepath in glob.glob('ui/tests/e2e/test-data.ts'):
    with open(filepath, 'r') as f:
        content = f.read()

    content = content.replace("import { ServiceTemplate } from '@proto/config/v1/service_template';", "import { ServiceTemplate } from '../../src/lib/client';")

    with open(filepath, 'w') as f:
        f.write(content)
