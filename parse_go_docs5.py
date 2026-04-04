import os
import re

def main():
    missing_docs = 0
    total_funcs = 0
    for root, dirs, files in os.walk('server'):
        if "test" in root: continue
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go'):
                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()

                # regex to capture the docblock and the func signature
                matches = re.finditer(r'(?:\/\/[^\n]*\n)*func (?:\([^\)]+\)\s+)?([A-Z][a-zA-Z0-9_]*)\(', content)
                for match in matches:
                    total_funcs += 1
                    func_block = match.group(0)
                    if "Summary:" not in func_block:
                        missing_docs += 1
                        print(f"{filepath}: {match.group(1)}")

                # Also check type structs and interfaces
                matches = re.finditer(r'(?:\/\/[^\n]*\n)*type ([A-Z][a-zA-Z0-9_]*)\s+(?:struct|interface)', content)
                for match in matches:
                    total_funcs += 1
                    func_block = match.group(0)
                    if "Summary:" not in func_block:
                        missing_docs += 1
                        print(f"{filepath}: {match.group(1)}")

                # Also check exported consts and vars
                # ...

    print(f"Missing docs in {missing_docs} out of {total_funcs} exported items in server")

main()
