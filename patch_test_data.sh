#!/bin/bash

cat << 'INNER_EOF' > patch_test_data.py
import re

with open('ui/tests/e2e/test-data.ts', 'r') as f:
    content = f.read()

content = content.replace('].map((template) => ServiceTemplate.toJSON(ServiceTemplate.fromJSON(template)));', '];')
content = content.replace('].map((user) => User.toJSON(User.fromJSON(user)));', '];')

with open('ui/tests/e2e/test-data.ts', 'w') as f:
    f.write(content)
INNER_EOF
python3 patch_test_data.py
