import json
import re
import os

def infer_meaning(name, kind, params=None, is_error=False):
    # Try to make a meaningful, non-empty calorie description
    name = name.lower()

    # Types
    if kind == "type":
        if "config" in name:
            return f"configuration parameters for the {name.replace('config', '')} component."
        if "request" in name:
            return "the structure of an incoming request."
        if "response" in name:
            return "the structure of an outgoing response."
        if "manager" in name:
            return "the central registry and orchestrator for this domain."
        if "interface" in name:
            return "the contract that implementations must fulfill."
        return "the core data structure for this domain entity."

    # Funcs / Methods summary
    if kind in ["func", "method"]:
        action = "Processes"
        if name.startswith("get"): action = "Retrieves"
        elif name.startswith("set"): action = "Updates"
        elif name.startswith("new"): action = "Initializes and returns a new"
        elif name.startswith("create"): action = "Instantiates"
        elif name.startswith("delete"): action = "Removes"
        elif name.startswith("handle"): action = "Processes incoming"
        elif name.startswith("is"): action = "Checks whether"
        elif name.startswith("validate"): action = "Verifies the integrity of"
        elif name.startswith("parse"): action = "Extracts data from"

        target = name
        for prefix in ["get", "set", "new", "create", "delete", "handle", "is", "validate", "parse"]:
            if name.startswith(prefix):
                target = name[len(prefix):]
                break

        target = target or "the operation"
        # add spaces to camelCase target
        target = re.sub(r"([A-Z])", r" \1", target).strip().lower()

        return f"{action} {target}."

    # Parameters
    if kind == "param":
        if name in ["ctx", "context"]: return "The cancellation and deadline context."
        if name in ["id"]: return "The unique identifier."
        if name in ["req", "request"]: return "The incoming request payload."
        if name in ["res", "response"]: return "The outgoing response payload."
        if name in ["err", "error"]: return "Any error encountered during execution."
        if name in ["cfg", "config"]: return "The configuration settings."
        if "name" in name: return "The human-readable or system name."
        if "url" in name: return "The endpoint address."
        if "data" in name: return "The raw payload."

        # fallback based on type
        if "string" in params: return f"The textual representation of {name}."
        if "int" in params or "float" in params: return f"The numeric value for {name}."
        if "bool" in params: return f"A flag indicating whether {name} is enabled."

        return f"The provided {name} data."

    # Returns
    if kind == "return":
        if is_error: return "An error if the execution fails, otherwise nil."
        if "string" in params: return "The resulting text."
        if "int" in params or "float" in params: return "The calculated numeric value."
        if "bool" in params: return "True if successful or valid, false otherwise."
        return "The resulting object or data structure."

    return "the specified value."

def generate_gold_standard_doc(node):
    lines = []

    # Existing doc handling
    existing_doc = node.get("doc", "").strip()
    summary = ""

    if existing_doc:
        # Check if existing doc already has Gold Standard elements
        if "Summary:" in existing_doc:
            return "" # Don't touch it if it's already compliant

        first_line = existing_doc.split('\n')[0].strip()
        lines.append(f"// {first_line}")
        lines.append("//")
        lines.append(f"// Summary: {first_line}")
    else:
        name = node["name"]
        summary = infer_meaning(name, node["kind"])

        # Capitalize first letter of name for the standard go comment
        if node["kind"] in ["func", "method", "type", "const", "var"]:
            lines.append(f"// {name} {summary}")
        lines.append("//")
        # Capitalize summary for Summary:
        summary_cap = summary[0].upper() + summary[1:] if summary else "Executes the operation."
        lines.append(f"// Summary: {summary_cap}")

    if node["kind"] in ["func", "method"]:
        lines.append("//")

        if node.get("params"):
            lines.append("// Parameters:")
            for p in node["params"]:
                pname = p["name"] if p["name"] else "unnamed"
                ptype = p["type"].replace('\n', ' ').replace('\t', '')
                desc = infer_meaning(pname, "param", ptype)
                lines.append(f"//   - {pname} ({ptype}): {desc}")
        else:
            lines.append("// Parameters:")
            lines.append("//   - None.")

        lines.append("//")

        if node.get("results"):
            lines.append("// Returns:")
            for r in node["results"]:
                rname = r["name"] if r["name"] else "unnamed"
                rtype = r["type"].replace('\n', ' ').replace('\t', '')
                desc = infer_meaning(rname, "return", rtype, is_error=(rtype=="error"))
                if rname != "unnamed":
                    lines.append(f"//   - {rname} ({rtype}): {desc}")
                else:
                    lines.append(f"//   - {rtype}: {desc}")
        else:
            lines.append("// Returns:")
            lines.append("//   - None.")

        lines.append("//")

        has_err = any(r["type"].replace('\n', ' ').replace('\t', '') == "error" for r in node.get("results", []))
        if has_err:
            lines.append("// Errors:")
            lines.append("//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.")
        else:
            lines.append("// Errors:")
            lines.append("//   - None.")

        lines.append("//")
        lines.append("// Side Effects:")
        lines.append("//   - May modify internal state or perform external network calls.")

    return "\n".join(lines) + "\n"

def process_file(file_path, file_nodes):
    with open(file_path, "r") as f:
        lines = f.readlines()

    for node in sorted(file_nodes, key=lambda x: x["pos"], reverse=True):
        new_doc = generate_gold_standard_doc(node)
        if not new_doc: continue # skip if already compliant

        if node.get("doc_start", 0) > 0 and node.get("doc_end", 0) >= node.get("doc_start", 0):
            start_idx = node["doc_start"] - 1
            end_idx = node["doc_end"]

            if any("package " in lines[i] or "import " in lines[i] for i in range(start_idx, end_idx)):
                insert_idx = node["pos"] - 1
                lines.insert(insert_idx, new_doc)
            else:
                lines[start_idx:end_idx] = []
                lines.insert(start_idx, new_doc)
        else:
            insert_idx = node["pos"] - 1
            lines.insert(insert_idx, new_doc)

    with open(file_path, "w") as f:
        f.writelines(lines)

with open('nodes.json', 'r') as f:
    nodes = json.load(f)

from collections import defaultdict
files = defaultdict(list)
for node in nodes:
    # ONLY apply to server/pkg to be safe from Bazel and proto and UI issues
    if node['file'].startswith('server/pkg') and node['file'].endswith('.go'):
        # also ignore anything in proto
        if "proto" not in node['file']:
            files[node['file']].append(node)

for file_path, file_nodes in files.items():
    process_file(file_path, file_nodes)
print("Done.")
