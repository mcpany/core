with open('server/roadmap.md', 'r') as f:
    lines = f.readlines()

new_lines = []
for i, line in enumerate(lines):
    # Line numbers are 1-based, so 504 is index 503
    if i == 503 or i == 506:
        new_lines.append(line.replace('*  ', '* '))
    else:
        new_lines.append(line)

with open('server/roadmap.md', 'w') as f:
    f.writelines(new_lines)
