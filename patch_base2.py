import re

filename = 'server/pkg/tool/base.go'
with open(filename, 'r') as f:
    content = f.read()

if "context" not in content:
    content = content.replace('import (\n\t"sync"', 'import (\n\t"context"\n\t"sync"')

with open(filename, 'w') as f:
    f.write(content)
