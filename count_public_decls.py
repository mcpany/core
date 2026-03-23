import os
import re

def count_decls(folder):
    count = 0
    for root, dirs, files in os.walk(folder):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go'):
                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()
                # find public funcs
                funcs = re.findall(r'^func\s+([A-Z]\w*)', content, re.MULTILINE)
                # find public methods
                methods = re.findall(r'^func\s+\([^)]+\)\s+([A-Z]\w*)', content, re.MULTILINE)
                # find public types
                types = re.findall(r'^type\s+([A-Z]\w*)', content, re.MULTILINE)

                count += len(funcs) + len(methods) + len(types)
    return count

print("Total public declarations:", count_decls("server/pkg"))
