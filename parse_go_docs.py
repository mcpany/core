import os
import re

def main():
    missing_docs = 0
    total_funcs = 0
    for root, dirs, files in os.walk('server/pkg'):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go'):
                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()

                # Match exported functions and methods
                # e.g., func MyFunc or func (s *Struct) MyFunc
                matches = re.finditer(r'^(?:\/\/(?:[^\n]*)\n)*func (?:\([^\)]+\)\s+)?([A-Z][a-zA-Z0-9_]*)\(', content, re.MULTILINE)
                for match in matches:
                    total_funcs += 1
                    func_block = match.group(0)
                    if "Summary:" not in func_block:
                        missing_docs += 1
                        # print(f"{filepath}: {match.group(1)}")
    print(f"Missing docs in {missing_docs} out of {total_funcs} exported functions in server/pkg")

main()
