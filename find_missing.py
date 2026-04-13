import os
import re

def main():
    missing_docs = []

    # regex for public funcs, types, interfaces, constants, vars
    # A bit naive but should work for top level declarations
    decl_regex = re.compile(r'^(func|type|const|var) (\([^)]+\)\s+)?([A-Z][a-zA-Z0-9_]*)')

    for root, dirs, files in os.walk('server/pkg'):
        for f in files:
            if f.endswith('.go') and not f.endswith('_test.go') and not f.endswith('.pb.go'):
                path = os.path.join(root, f)
                with open(path, 'r') as file:
                    content = file.read()

                lines = content.split('\n')
                for i, line in enumerate(lines):
                    # For const and var, they might be in a block `const (`
                    # Let's simplify and just focus on functions, types, and single line declarations for now.
                    # Or we can just look for things starting with `func`, `type`, `const`, `var`
                    m = decl_regex.match(line)
                    if m:
                        decl_type = m.group(1)
                        name = m.group(3)

                        has_doc = False
                        j = i - 1
                        doc_block = []
                        while j >= 0 and lines[j].startswith('//'):
                            doc_block.insert(0, lines[j])
                            j -= 1

                        doc_text = '\n'.join(doc_block)

                        if decl_type == 'func':
                            if 'Summary:' not in doc_text or 'Returns:' not in doc_text:
                                missing_docs.append(f"{path}:{i+1} - {line}")
                        elif decl_type in ('type', 'const', 'var'):
                            if 'Summary:' not in doc_text:
                                missing_docs.append(f"{path}:{i+1} - {line}")

    with open('missing_docs.txt', 'w') as f:
        f.write('\n'.join(missing_docs))

    print(f"Total Missing: {len(missing_docs)}")

if __name__ == '__main__':
    main()
