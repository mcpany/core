import re

with open('ui/src/tests/components/stacks/stack-editor.test.tsx', 'r') as f:
    content = f.read()

content = content.replace('import { render, screen, fireEvent } from \'@testing-library/react\';', 'import { render, screen } from \'@testing-library/react\';')
with open('ui/src/tests/components/stacks/stack-editor.test.tsx', 'w') as f:
    f.write(content)


with open('ui/src/tests/components/secrets-manager.test.tsx', 'r') as f:
    content = f.read()

content = content.replace('import { render, screen, waitFor } from \'@testing-library/react\';', 'import { render, screen } from \'@testing-library/react\';')
content = content.replace('import { render, screen, waitFor } from \'@testing-library/react\'', 'import { render, screen } from \'@testing-library/react\'')

with open('ui/src/tests/components/secrets-manager.test.tsx', 'w') as f:
    f.write(content)
