import os
for d in ['server/pkg', 'server/cmd', 'server/api']:
    for root, _, files in os.walk(d):
        for f in files:
            if f.endswith('.go') and not f.endswith('_test.go'):
                path = os.path.join(root, f)
                with open(path, 'r') as file:
                    content = file.read()

                if 'package ' not in content:
                    pkg_name = os.path.basename(os.path.dirname(path))
                    print(f"Fixing {path} with package {pkg_name}")

                    # Find first type or var or func or import and put package before it
                    lines = content.split('\n')
                    for i, l in enumerate(lines):
                        if l.startswith('import ') or l.startswith('type ') or l.startswith('func ') or l.startswith('var '):
                            lines.insert(i, f"package {pkg_name}\n")
                            break
                    with open(path, 'w') as file:
                        file.write('\n'.join(lines))
