import re
with open('server/tests/framework/aitool.go', 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    if line.startswith('ry: ') or line.startswith(' Re-exporting ') or line.startswith('ummary: '):
        continue
    new_lines.append(line)

with open('server/tests/framework/aitool.go', 'w') as f:
    f.writelines(new_lines)
