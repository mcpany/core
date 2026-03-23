import os
import re

def process_file(filepath):
    # Only target pkg for public API
    if "test" in filepath or "examples" in filepath or "mock" in filepath:
        return 0
    if not filepath.startswith('server/pkg/'):
        return 0

    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    changes = 0
    i = 0
    while i < len(lines):
        line = lines[i]

        match = re.match(r'^func\s+([A-Z]\w*)\s*\((.*?)\)\s*(.*?)\{', line)
        method_match = re.match(r'^func\s+\([^)]+\)\s+([A-Z]\w*)\s*\((.*?)\)\s*(.*?)\{', line)

        type_match = re.match(r'^type\s+([A-Z]\w*)\s+', line)
        var_match = re.match(r'^(var|const)\s+([A-Z]\w*)', line)

        name = None
        params_str = ""
        returns_str = ""
        is_func = False

        if match:
            name, params_str, returns_str = match.groups()
            is_func = True
        elif method_match:
            name, params_str, returns_str = method_match.groups()
            is_func = True
        elif type_match:
            name = type_match.group(1)
        elif var_match:
            name = var_match.group(2)

        if name and not name.startswith('Test'):
            doc_start = -1
            for j in range(i - 1, -1, -1):
                if lines[j].strip().startswith('//'):
                    doc_start = j
                elif not lines[j].strip():
                    continue
                else:
                    break

            if doc_start == -1:
                # Missing completely
                doc_lines = [f"// {name} ...\n", "//\n"]
                if is_func:
                    doc_lines.append(f"// Summary: Handles {name}.\n//\n")

                    # Params
                    params = [p.strip() for p in params_str.split(',') if p.strip() and p.strip() != '']
                    actual_params = []
                    for p in params:
                        parts = p.split()
                        if len(parts) >= 2:
                            p_name = parts[0]
                            p_type = " ".join(parts[1:])
                            actual_params.append((p_name, p_type))
                        elif len(parts) == 1:
                            actual_params.append((parts[0], "interface{}"))

                    if actual_params:
                        doc_lines.append("// Parameters:\n")
                        for p_name, p_type in actual_params:
                            if p_name == "ctx":
                                doc_lines.append(f"//   - {p_name} ({p_type}): Execution context.\n")
                            else:
                                doc_lines.append(f"//   - {p_name} ({p_type}): The {p_name} argument.\n")
                        doc_lines.append("//\n")

                    # Returns
                    ret = returns_str.strip()
                    if ret and ret != "{":
                        doc_lines.append("// Returns:\n")
                        doc_lines.append(f"//   - {ret}: The output.\n")
                        doc_lines.append("//\n")

                    if "error" in ret:
                        doc_lines.append("// Errors:\n")
                        doc_lines.append(f"//   - Returns an error if {name} encounters a failure.\n")
                        doc_lines.append("//\n")
                else:
                    doc_lines.append(f"// Summary: Defines the {name} entity.\n//\n")

                for l in reversed(doc_lines):
                    lines.insert(i, l)
                i += len(doc_lines)
                changes += 1

            else:
                # Analyze existing doc block to add specifically missing chunks
                doc_block = "".join(lines[doc_start:i])
                insertions = []

                # Check for Summary:
                if "Summary:" not in doc_block:
                    if is_func:
                        insertions.append((doc_start, f"// Summary: Executes {name}.\n//\n"))
                    else:
                        insertions.append((doc_start, f"// Summary: Defines {name}.\n//\n"))

                if is_func:
                    params = [p.strip() for p in params_str.split(',') if p.strip() and p.strip() != '']
                    actual_params = []
                    for p in params:
                        parts = p.split()
                        if len(parts) >= 2:
                            p_name = parts[0]
                            p_type = " ".join(parts[1:])
                            actual_params.append((p_name, p_type))
                        elif len(parts) == 1:
                            actual_params.append((parts[0], "value"))

                    if actual_params and "Parameters:" not in doc_block:
                        param_lines = ["// Parameters:\n"]
                        for p_name, p_type in actual_params:
                            if p_name != "ctx" and p_name != "err":
                                param_lines.append(f"//   - {p_name} ({p_type}): Input for {p_name}.\n")
                            else:
                                param_lines.append(f"//   - {p_name} ({p_type}): Standard context.\n")
                        param_lines.append("//\n")

                        idx = i
                        for j in range(doc_start, i):
                            if "Returns:" in lines[j] or "Errors:" in lines[j] or "Side Effects:" in lines[j]:
                                idx = j
                                break
                        insertions.append((idx, "".join(param_lines)))

                    ret = returns_str.strip()
                    if ret and ret != "{" and "Returns:" not in doc_block:
                        ret_lines = ["// Returns:\n", f"//   - {ret}: The output.\n", "//\n"]
                        idx = i
                        for j in range(doc_start, i):
                            if "Errors:" in lines[j] or "Side Effects:" in lines[j]:
                                idx = j
                                break
                        insertions.append((idx, "".join(ret_lines)))

                    if "error" in ret and "Errors:" not in doc_block:
                        err_lines = ["// Errors:\n", f"//   - error on {name} failure.\n", "//\n"]
                        idx = i
                        for j in range(doc_start, i):
                            if "Side Effects:" in lines[j]:
                                idx = j
                                break
                        insertions.append((idx, "".join(err_lines)))

                if insertions:
                    insertions.sort(key=lambda x: x[0], reverse=True)
                    for idx, text in insertions:
                        for l in reversed(text.splitlines(True)):
                            lines.insert(idx, l)
                            i += 1
                    changes += 1

        i += 1

    if changes > 0:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.writelines(lines)
    return changes

def main():
    total = 0
    for root, dirs, files in os.walk('server/pkg/'):
        for f in files:
            if f.endswith('.go') and not f.endswith('_test.go') and not f.endswith('.pb.go') and 'mock_' not in f:
                filepath = os.path.join(root, f)
                c = process_file(filepath)
                if c > 0:
                    total += c
                    print(f"Fixed {c} issues in {filepath}")
    print(f"Total: {total}")

if __name__ == '__main__':
    main()
