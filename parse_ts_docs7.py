import os
import re
import json

def main():
    missing_docs = []

    for root, dirs, files in os.walk('ui/src'):
        for file in files:
            if file.endswith('.ts') or file.endswith('.tsx'):
                if file.endswith('.d.ts'): continue

                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()

                # Match exported components, functions, interfaces, types, consts
                matches = re.finditer(r'(?:(\/\*\*[\s\S]*?\*\/)[\s\n]*)?(export\s+(?:default\s+)?(?:const|function|class|interface|type)\s+([A-Z][a-zA-Z0-9_]*|))', content)
                for match in matches:
                    doc = match.group(1)
                    export_stmt = match.group(2)
                    name = match.group(3)

                    if not name and 'default' in export_stmt:
                        continue
                    if not name:
                        continue

                    # Check if doc contains "Intent:" OR "Summary:" (user's prompt prefers "Summary:")
                    if not doc or ("Intent:" not in doc and "Summary:" not in doc):
                        missing_docs.append({
                            'file': filepath,
                            'name': name,
                            'type': export_stmt.split()[1] # const/function/class/interface/type
                        })

    print(f"Missing docs in {len(missing_docs)} exported items in ui/src")
    with open('missing_ts_docs.json', 'w') as f:
        json.dump(missing_docs, f, indent=2)

main()
