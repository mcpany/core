import os
import re
import sys

def find_undocumented_go(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    undocumented = []

    in_comment_block = False

    for i, line in enumerate(lines):
        stripped_line = line.strip()

        # Check if line is a comment
        if stripped_line.startswith('//') or stripped_line.startswith('/*') or stripped_line.endswith('*/') or in_comment_block:
            if '/*' in stripped_line and '*/' not in stripped_line:
                in_comment_block = True
            if '*/' in stripped_line:
                in_comment_block = False
            continue

        # Check for exported function
        match = re.match(r'^func\s+([A-Z][a-zA-Z0-9_]*)', stripped_line)
        if match:
            prev_line = lines[i-1].strip() if i > 0 else ""
            if i == 0 or not (prev_line.startswith('//') or prev_line.endswith('*/')):
                 undocumented.append((i + 1, "func " + match.group(1)))
            continue

        # Check for exported method
        # func (r Receiver) MethodName(...)
        match = re.match(r'^func\s+\([^)]+\)\s+([A-Z][a-zA-Z0-9_]*)', stripped_line)
        if match:
            prev_line = lines[i-1].strip() if i > 0 else ""
            if i == 0 or not (prev_line.startswith('//') or prev_line.endswith('*/')):
                 undocumented.append((i + 1, "method " + match.group(1)))
            continue

        # Check for exported type
        match = re.match(r'^type\s+([A-Z][a-zA-Z0-9_]*)', stripped_line)
        if match:
             prev_line = lines[i-1].strip() if i > 0 else ""
             if i == 0 or not (prev_line.startswith('//') or prev_line.endswith('*/')):
                 undocumented.append((i + 1, "type " + match.group(1)))
             continue

        # Check for exported const/var
        match = re.match(r'^(const|var)\s+([A-Z][a-zA-Z0-9_]*)', stripped_line)
        if match:
             prev_line = lines[i-1].strip() if i > 0 else ""
             if i == 0 or not (prev_line.startswith('//') or prev_line.endswith('*/')):
                 undocumented.append((i + 1, match.group(1) + " " + match.group(2)))
             continue

    return undocumented

def find_undocumented_ts(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    undocumented = []

    in_comment_block = False

    for i, line in enumerate(lines):
        stripped_line = line.strip()

        # Check if line is a comment
        if stripped_line.startswith('//') or stripped_line.startswith('/*') or stripped_line.endswith('*/') or in_comment_block:
            if '/*' in stripped_line and '*/' not in stripped_line:
                in_comment_block = True
            if '*/' in stripped_line:
                in_comment_block = False
            continue

        # Check for exported symbols
        if stripped_line.startswith('export '):
            # Check if it has a comment before it
             prev_line = lines[i-1].strip() if i > 0 else ""
             if i == 0 or not (prev_line.startswith('//') or prev_line.endswith('*/')):
                 # Try to extract name
                 name = "unknown"

                 parts = stripped_line.split()
                 if len(parts) >= 3 and parts[1] in ['const', 'function', 'class', 'interface', 'type', 'enum']:
                     name = parts[2].split('(')[0].split(':')[0].split('=')[0].split('<')[0]
                 elif len(parts) >= 3 and parts[1] == 'default':
                     name = 'default'

                 if name != 'default':
                    undocumented.append((i + 1, stripped_line.split('{')[0].strip()))

    return undocumented

def main():
    if len(sys.argv) < 2:
        print("Usage: python find_undocumented.py <directory>")
        sys.exit(1)

    root_path = sys.argv[1]

    if os.path.isfile(root_path):
        if root_path.endswith('.go'):
            undoc = find_undocumented_go(root_path)
            if undoc:
                print(f"\nFile: {root_path}")
                for line_num, sig in undoc:
                    print(f"  Line {line_num}: {sig}")
        elif root_path.endswith('.ts') or root_path.endswith('.tsx'):
            undoc = find_undocumented_ts(root_path)
            if undoc:
                print(f"\nFile: {root_path}")
                for line_num, sig in undoc:
                    print(f"  Line {line_num}: {sig}")
    else:
        for root, dirs, files in os.walk(root_path):
            if 'node_modules' in dirs:
                dirs.remove('node_modules')
            if 'vendor' in dirs:
                dirs.remove('vendor')
            if '.git' in dirs:
                dirs.remove('.git')
            if 'build' in dirs:
                dirs.remove('build')

            for file in files:
                filepath = os.path.join(root, file)
                if file.endswith('.go') and not file.endswith('_test.go'):
                    undoc = find_undocumented_go(filepath)
                    if undoc:
                        print(f"\nFile: {filepath}")
                        for line_num, sig in undoc:
                            print(f"  Line {line_num}: {sig}")

                elif (file.endswith('.ts') or file.endswith('.tsx')) and not file.endswith('.test.ts') and not file.endswith('.test.tsx') and not file.endswith('.spec.ts'):
                    undoc = find_undocumented_ts(filepath)
                    if undoc:
                        print(f"\nFile: {filepath}")
                        for line_num, sig in undoc:
                            print(f"  Line {line_num}: {sig}")

if __name__ == "__main__":
    main()
