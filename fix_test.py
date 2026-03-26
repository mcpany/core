import os
import re

files_to_fix = [
    "server/pkg/command/command_test.go",
]

for file_path in files_to_fix:
    with open(file_path, 'r') as f:
        content = f.read()

    # We want to uncomment skipped tests or remove the skip line
    content = re.sub(r'(?m)^\s*t\.Skip\(.*?$\n?', '', content)

    with open(file_path, 'w') as f:
        f.write(content)
