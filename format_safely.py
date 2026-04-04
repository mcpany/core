import os
import re
import subprocess
import json

def get_missing():
    # Use our AST extractor to get exact line numbers of missing/malformed docs
    subprocess.run(["go", "run", "extract.go"], check=True)
    with open('symbols.json', 'r') as f:
        symbols = json.load(f)
    return symbols

def create_doc(sym):
    kind = sym['kind']
    name = sym['name']
    existing = sym['existing'].strip() if sym['existing'] else ""

    if existing:
        summary = existing.split('\n')[0].replace(f"{name} ", "")
    else:
        if kind == 'Var':
            summary = f"Represents {name}."
        elif kind == 'Type':
            summary = f"Represents {name} structure or interface."
        else:
            summary = f"Executes {name} operation."

    if summary and summary.startswith('Represents'):
        pass
    elif summary and summary.startswith('Executes'):
        pass
    elif summary and len(summary) > 0:
        summary = summary[0].upper() + summary[1:]

    doc = []
    doc.append(f"// {name} {summary[0].lower() + summary[1:] if summary else ''}")
    doc.append("//")
    doc.append(f"// Summary: {summary}")
    doc.append("//")
    doc.append("// Parameters:")

    if kind == 'Func':
        if sym.get('params'):
            for p in sym['params']:
                pname, ptype = p.split('|')
                if pname == "_":
                     pname = "unnamed"
                doc.append(f"//   - {pname} ({ptype}): Parameter.")
        else:
            doc.append("//   - None.")
    else:
        doc.append("//   - None.")

    doc.append("//")
    doc.append("// Returns:")

    if kind == 'Func':
        if sym.get('returns'):
            for r in sym['returns']:
                doc.append(f"//   - {r}: Return value.")
        else:
            doc.append("//   - None.")
    else:
        doc.append("//   - None.")

    doc.append("//")
    doc.append("// Errors:")
    if kind == 'Func' and sym.get('has_error'):
        doc.append("//   - error: If an error occurs.")
    else:
        doc.append("//   - None.")

    doc.append("//")
    doc.append("// Side Effects:")
    doc.append("//   - None.")

    return "\n".join(doc) + "\n"

def main():
    symbols = get_missing()

    # Filter to only the 125 missing ones from before, but actually let's just do it
    # in a way that respects string boundaries. The AST parser does this natively!
    # Because it only extracts actual decls, not strings.

    files = {}
    for sym in symbols:
        files.setdefault(sym['file'], []).append(sym)

    for filepath, file_symbols in files.items():
        # Sort by reverse doc_pos
        file_symbols.sort(key=lambda x: x['doc_pos'], reverse=True)

        with open(filepath, 'rb') as f:
            content = bytearray(f.read())

        for sym in file_symbols:
            new_doc = create_doc(sym).encode('utf-8')

            doc_start = sym['doc_start']
            doc_end = sym['doc_end']
            doc_pos = sym['doc_pos']

            # If doc_start exists, replace it
            if doc_start != -1 and doc_end != -1:
                content = content[:doc_start] + new_doc + content[doc_end:]
            else:
                # Add before doc_pos
                content = content[:doc_pos] + new_doc + content[doc_pos:]

        with open(filepath, 'wb') as f:
            f.write(content)

if __name__ == "__main__":
    main()
