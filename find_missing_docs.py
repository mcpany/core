import os
import re

def check_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    missing = []
    # Regex for exported symbols: Start of line, type/func/var/const, space, Uppercase letter
    # This is a heuristic.
    exported_pattern = re.compile(r'^(func|type|var|const)\s+([A-Z]\w*)')

    # Method pattern: func (r *Receiver) MethodName
    method_pattern = re.compile(r'^func\s+\([^)]+\)\s+([A-Z]\w*)')

    for i, line in enumerate(lines):
        symbol = None
        match = exported_pattern.match(line)
        if match:
            symbol = match.group(2)
        else:
            match = method_pattern.match(line)
            if match:
                symbol = match.group(1)

        if symbol:
            # Check previous lines for comments
            doc_lines = []
            j = i - 1
            while j >= 0:
                prev = lines[j].strip()
                if prev.startswith('//'):
                    if not prev.startswith('//go:'):
                        doc_lines.insert(0, prev)
                    if prev == '': # break on empty comment line? No, // usually has content. empty line is just empty string.
                        pass
                elif prev == '':
                    break
                else:
                    break
                j -= 1

            # Verify if it has structured docs
            if not doc_lines:
                missing.append((i + 1, symbol, "Missing doc"))
            else:
                doc_text = "\n".join(doc_lines)
                # Heuristic: Check for "Parameters:" or "Returns:" if it's a function
                if line.startswith('func'):
                    if "Parameters:" not in doc_text and "Returns:" not in doc_text:
                         # Skip simple getters/setters? Or simple constructors?
                         # The prompt says "100% coverage... inject a structured docstring".
                         # Let's flag it.
                         missing.append((i + 1, symbol, "Non-compliant doc"))

    return missing

def scan_dir(root_dir):
    for root, dirs, files in os.walk(root_dir):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go') and 'vendor' not in root:
                path = os.path.join(root, file)
                missing = check_file(path)
                if missing:
                    print(f"File: {path}")
                    for line, symbol, reason in missing:
                        print(f"  Line {line}: {symbol} ({reason})")

if __name__ == '__main__':
    scan_dir('server/pkg')
    scan_dir('server/cmd')
