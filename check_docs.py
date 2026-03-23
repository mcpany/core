import os
import re

def check_docs(folder):
    missing_docs = 0
    total_decls = 0

    for root, dirs, files in os.walk(folder):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go'):
                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()

                # regex to find exported functions/methods/types
                # This is a bit simplistic, but let's see
                decls = re.finditer(r'(?:^\s*//.*$\n)*^func\s+(?:(?:\([^)]+\)\s+)|)([A-Z]\w*)', content, re.MULTILINE)

                for match in decls:
                    total_decls += 1
                    func_str = match.group(0)
                    if '// Summary:' not in func_str:
                        missing_docs += 1

    return total_decls, missing_docs

t, m = check_docs("server/pkg")
print(f"Total decls: {t}, Missing docs: {m}")
