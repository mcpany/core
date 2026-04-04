import os
import re

def main():
    missing_docs = []
    total_funcs = 0
    for root, dirs, files in os.walk('server'):
        if "test" in root: continue
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go'):
                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()

                # Try to extract the whole doc block
                # regex to capture the docblock and the func signature
                matches = re.finditer(r'((?:\/\/[^\n]*\n)*)func (?:\([^\)]+\)\s+)?([A-Z][a-zA-Z0-9_]*)\(', content)
                for match in matches:
                    total_funcs += 1
                    doc_block = match.group(1)
                    if "Summary:" not in doc_block:
                        missing_docs.append((filepath, "func", match.group(2)))

                # Also check type structs and interfaces
                matches = re.finditer(r'((?:\/\/[^\n]*\n)*)type ([A-Z][a-zA-Z0-9_]*)\s+(?:struct|interface)', content)
                for match in matches:
                    total_funcs += 1
                    doc_block = match.group(1)
                    if "Summary:" not in doc_block:
                        missing_docs.append((filepath, "type", match.group(2)))

    print(f"Missing docs in {len(missing_docs)} out of {total_funcs} exported items in server")
    for doc in missing_docs:
        print(doc)

main()
