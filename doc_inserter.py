import os
import re

def expand_name(name, kind):
    words = re.findall(r'[A-Z]?[a-z]+|[A-Z]+(?=[A-Z]|$)', name)
    if not words:
        return name

    lower_words = [w.lower() if w.lower() not in ["id", "url", "json", "api", "mcp", "http", "grpc"] else w.upper() for w in words]
    lower_words[0] = lower_words[0].capitalize()

    if kind == "type":
        return f"Represents the {' '.join(lower_words).lower()}."
    elif kind == "var":
        return f"Global variable holding {' '.join(lower_words).lower()}."
    elif kind == "const":
        return f"Constant value defining {' '.join(lower_words).lower()}."

    verb = lower_words[0].lower()
    rest = " ".join(lower_words[1:])

    if verb in ["get", "set", "create", "update", "delete", "add", "remove", "parse", "load", "save", "start", "stop", "run", "execute", "build", "generate", "find", "search", "read", "write"]:
        action = verb + "s"
    elif verb.endswith("e"):
        action = verb + "s"
    elif verb.endswith("y"):
        action = verb[:-1] + "ies"
    else:
        action = verb + "s"

    action_phrase = action.capitalize() + (" " + rest if rest else "")
    return action_phrase + "."

def build_doc(name, kind, params=[], returns=[]):
    action = expand_name(name, kind)

    doc = f"// {name} {action[0].lower() + action[1:]}\n"
    if kind in ["type", "var", "const"]:
        doc += f"//\n// Summary: {action}\n"
        return doc

    doc += f"//\n// Summary: {action}\n"
    doc += f"//\n// Parameters:\n"
    if not params:
        doc += f"//   - None.\n"
    else:
        for p in params:
            doc += f"//   - {p} (any): The {p} argument.\n"

    doc += f"//\n// Returns:\n"
    if not returns:
        doc += f"//   - None.\n"
    else:
        doc += f"//   - result (any): The output result.\n"

    doc += f"//\n// Errors:\n//   - Returns error upon failure.\n"
    doc += f"//\n// Side Effects:\n//   - Interacts with internal state.\n"
    return doc

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    out_lines = []

    i = 0
    while i < len(lines):
        line = lines[i]

        func_match = re.match(r'^func\s+(?:\([^\)]+\)\s+)?([A-Z][a-zA-Z0-9_]*)\s*\(', line)
        type_match = re.match(r'^type\s+([A-Z][a-zA-Z0-9_]*)\s+(struct|interface)', line)
        var_match = re.match(r'^(?:var|const)\s+(?:\(\s*)?([A-Z][a-zA-Z0-9_]*)\s*=', line)

        if func_match or type_match or var_match:
            if func_match:
                name = func_match.group(1)
                kind = "func"
            elif type_match:
                name = type_match.group(1)
                kind = "type"
            else:
                name = var_match.group(1)
                kind = line.split()[0]

            has_doc = False
            has_summary = False

            for j in range(i-1, -1, -1):
                if lines[j].strip().startswith("//"):
                    has_doc = True
                    if "Summary:" in lines[j]:
                        has_summary = True
                elif lines[j].strip() == "":
                    continue
                else:
                    break

            if not has_summary:
                doc_lines = build_doc(name, kind).split('\n')
                if has_doc:
                    structured_part = doc_lines[1:]
                    for dl in structured_part:
                        if dl: out_lines.append(dl)
                else:
                    for dl in doc_lines:
                        if dl: out_lines.append(dl)

        out_lines.append(line)

        if type_match and type_match.group(2) == 'interface':
            in_interface = True

        if 'in_interface' in locals() and in_interface and line.strip() == '}':
            in_interface = False

        if 'in_interface' in locals() and in_interface and re.match(r'^\s+([A-Z][a-zA-Z0-9_]*)\s*\(', line):
            method_match = re.match(r'^\s+([A-Z][a-zA-Z0-9_]*)\s*\(', line)
            name = method_match.group(1)

            has_summary = False
            for j in range(i-1, -1, -1):
                if lines[j].strip().startswith("//"):
                    if "Summary:" in lines[j]:
                        has_summary = True
                else:
                    break

            if not has_summary:
                doc_lines = build_doc(name, "method").split('\n')
                structured_part = doc_lines[1:]

                has_doc = lines[i-1].strip().startswith("//")

                if has_doc:
                    for dl in structured_part:
                        if dl: out_lines.insert(-1, "\t" + dl)
                else:
                    for dl in doc_lines:
                        if dl: out_lines.insert(-1, "\t" + dl)

        i += 1

    with open(filepath, 'w') as f:
        f.write('\n'.join(out_lines))

for root, _, files in os.walk('server'):
    if 'vendor' in root or 'build' in root or 'tools' in root:
        continue
    for file in files:
        if file.endswith('.go') and not file.endswith('_test.go') and not file.endswith('.pb.go') and not file.endswith('.pb.gw.go'):
            process_file(os.path.join(root, file))
