import sys

def dedupe(filename):
    with open(filename, 'r') as f:
        lines = f.readlines()

    new_lines = []
    for line in lines:
        if not new_lines or line != new_lines[-1]:
            new_lines.append(line)

    with open(filename, 'w') as f:
        f.writelines(new_lines)

for arg in sys.argv[1:]:
    dedupe(arg)
