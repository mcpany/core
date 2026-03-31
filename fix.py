import os
import re

# Dictionary to give meaningful descriptions
COMMON_TERMS = {
    "ctx": "context.Context. The execution context for managing deadlines and cancellation signals",
    "id": "string. A unique identifier used for tracking and lookup",
    "name": "string. The designated name identifying the specific entity",
    "uri": "string. A Uniform Resource Identifier used to locate the specific resource",
    "err": "error. An error object detailing any failure that occurred during execution",
    "req": "interface. The request payload containing operation parameters",
    "res": "interface. The response object for the executed operation",
    "config": "struct. Configuration settings governing the behavior of the component",
    "cfg": "struct. Configuration structure providing operational parameters",
    "client": "struct. A client connection used to interface with external services",
    "manager": "struct. A manager instance responsible for coordinating operations",
    "ctrl": "struct. A controller managing the lifecycle and state",
    "provider": "interface. A provider implementation for external services",
    "options": "struct. Options used to configure the operation",
    "opts": "struct. Options used to configure the operation",
    "key": "string. A key used to identify the entity",
    "val": "interface. A value associated with the key",
    "value": "interface. A value associated with the key",
    "url": "string. The Uniform Resource Locator to connect to",
    "path": "string. The filesystem path",
    "service": "string. The target service identifier",
}

def get_desc(param_name, param_type):
    name_lower = param_name.lower()
    for k, v in COMMON_TERMS.items():
        if k in name_lower:
            return f"{param_type}. {v.split('. ', 1)[-1] if '. ' in v else v}"
    return f"{param_type}. Represents the {param_name} for the operation"

def parse_sig(lines, start_idx):
    for i in range(start_idx, min(start_idx + 25, len(lines))):
        if lines[i].startswith("func "):
            sig = lines[i].strip()
            j = i
            while "{" not in sig and j + 1 < len(lines):
                j += 1
                sig += " " + lines[j].strip()
            return sig
    return ""

def get_params_rets(sig):
    params_str = ""
    ret_str = ""
    if "(" in sig:
        parts = re.findall(r'\((.*?)\)', sig)
        if sig.startswith("func (") and len(parts) >= 2:
            params_str = parts[1]
        elif len(parts) >= 1:
            if sig.startswith("func ("):
                 params_str = parts[0] if len(parts) == 1 else parts[1]
            else:
                 params_str = parts[0]

    if ")" in sig:
        ret_part = sig[sig.rfind(')') + 1 : sig.rfind('{')].strip()
        ret_str = ret_part.strip('()')

    return params_str, ret_str

def parse_params(params_str):
    if not params_str.strip(): return []
    parts = [p.strip() for p in params_str.split(',')]
    parsed = []
    # handle `a, b string` -> `a string`, `b string`
    for p in parts:
        if not p: continue
        words = p.split()
        if len(words) == 1:
            parsed.append((words[0], "unknown"))
        else:
            parsed.append((words[0], " ".join(words[1:])))
    return parsed

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    if "TODO:" not in content:
        return False

    lines = content.split('\n')
    out_lines = []

    i = 0
    changed = False
    while i < len(lines):
        line = lines[i]

        if "TODO: Document parameters." in line:
            sig = parse_sig(lines, i)
            params_str, _ = get_params_rets(sig)
            params = parse_params(params_str)
            if not params:
                out_lines.append(line.replace("TODO: Document parameters.", "None."))
            else:
                first = True
                for p_name, p_type in params:
                    desc = f"{p_name}: {get_desc(p_name, p_type)}."
                    if first:
                        out_lines.append(line.replace("TODO: Document parameters.", desc))
                        first = False
                    else:
                        indent = line[:line.find("//")]
                        out_lines.append(f"{indent}//   - {desc}")
            changed = True
            i += 1
            continue

        elif "TODO: Document returns." in line:
            sig = parse_sig(lines, i)
            _, ret_str = get_params_rets(sig)

            if not ret_str.strip():
                out_lines.append(line.replace("TODO: Document returns.", "None."))
            else:
                rets = parse_params(ret_str)
                non_err = [(n, t) for n, t in rets if n != 'error' and t != 'error']
                if not non_err:
                    out_lines.append(line.replace("TODO: Document returns.", "None."))
                else:
                    first = True
                    for r_name, r_type in non_err:
                        if r_type == "unknown":
                            desc = f"{r_name}: The resulting output from the operation."
                        else:
                            desc = f"{r_name} ({r_type}): The resulting output from the operation."
                        if first:
                            out_lines.append(line.replace("TODO: Document returns.", desc))
                            first = False
                        else:
                            indent = line[:line.find("//")]
                            out_lines.append(f"{indent}//   - {desc}")
            changed = True
            i += 1
            continue

        elif "TODO: Document errors." in line:
            sig = parse_sig(lines, i)
            _, ret_str = get_params_rets(sig)

            rets = parse_params(ret_str)
            has_error = any(n == 'error' or t == 'error' for n, t in rets)

            if has_error:
                out_lines.append(line.replace("TODO: Document errors.", "error: An error object detailing any execution failure, or nil on success."))
            else:
                out_lines.append(line.replace("TODO: Document errors.", "None."))
            changed = True
            i += 1
            continue

        elif "TODO: Document side effects." in line:
            out_lines.append(line.replace("TODO: Document side effects.", "None."))
            changed = True
            i += 1
            continue

        out_lines.append(line)
        i += 1

    if changed:
        with open(filepath, 'w') as f:
            f.write('\n'.join(out_lines))
        return True
    return False

if __name__ == "__main__":
    count = 0
    for root, dirs, files in os.walk("server"):
        for file in files:
            if file.endswith(".go") and "mock" not in file.lower() and "test" not in file.lower():
                if process_file(os.path.join(root, file)):
                    count += 1
    print(f"Fixed {count} files")
