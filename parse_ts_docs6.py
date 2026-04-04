import os
import re

def main():
    missing_docs = 0
    total_funcs = 0
    missing_files = []

    for root, dirs, files in os.walk('ui/src'):
        for file in files:
            if file.endswith('.ts') or file.endswith('.tsx'):
                if file.endswith('.d.ts'): continue

                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()

                matches = re.finditer(r'(?:(\/\*\*[\s\S]*?\*\/)[\s\n]*)?(export\s+(?:default\s+)?(?:const|function|class|interface|type)\s+([A-Z][a-zA-Z0-9_]*|))', content)
                for match in matches:
                    doc = match.group(1)
                    export_stmt = match.group(2)
                    name = match.group(3)

                    if not name and 'default' in export_stmt:
                        continue
                    if not name:
                        continue

                    total_funcs += 1

                    if not doc or "Intent:" not in doc:
                        missing_docs += 1
                        missing_files.append((filepath, name))

    print(f"Missing docs in {missing_docs} out of {total_funcs} exported items in ui/src")

main()
