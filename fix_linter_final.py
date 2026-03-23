import re

path = "server/pkg/lint/linter.go"
with open(path, "r") as f:
    lines = f.readlines()

def format_doc(summary, params=None, returns=None, errors=None, side_effects=None):
    doc = [f"// Summary: {summary}\n", "//\n"]
    doc.append("// Parameters:\n")
    if params:
        for p in params: doc.append(f"//   - {p}\n")
    else: doc.append("//   - None.\n")
    doc.append("//\n// Returns:\n")
    if returns:
        for r in returns: doc.append(f"//   - {r}\n")
    else: doc.append("//   - None.\n")
    doc.append("//\n// Errors:\n")
    if errors:
        for e in errors: doc.append(f"//   - {e}\n")
    else: doc.append("//   - None.\n")
    doc.append("//\n// Side Effects:\n")
    if side_effects:
        for s in side_effects: doc.append(f"//   - {s}\n")
    else: doc.append("//   - None.\n")
    return doc

new_lines = []
i = 0
while i < len(lines):
    line = lines[i]
    if line.startswith("func (s Severity) String()"):
        new_lines.extend(format_doc("Executes String operation.", ["None."], ["string: The string representation."]))
        new_lines.append(line)
        i += 1
    elif line.startswith("func (r Result) String()"):
        new_lines.extend(format_doc("Executes String operation.", ["None."], ["string: A formatted string containing result details."]))
        new_lines.append(line)
        i += 1
    elif line.startswith("func NewLinter"):
        new_lines.extend(format_doc("Initializes NewLinter operation.", ["cfg (*configv1.McpAnyServerConfig): The configuration to analyze."], ["*Linter: A new Linter instance."]))
        new_lines.append(line)
        i += 1
    elif line.startswith("func (l *Linter) Run"):
        new_lines.extend(format_doc("Executes Run operation.", ["ctx (context.Context): The context for the operation."], ["[]Result: A slice of linting findings.", "error: Encounters a fatal issue."]))
        new_lines.append(line)
        i += 1
    elif line.startswith("func (l *Linter) checkPlainTextSecrets"):
        new_lines.extend(format_doc("Executes checkPlainTextSecrets operation.", ["None."], ["[]Result: A slice of findings."]))
        new_lines.append(line)
        i += 1
    elif line.startswith("func (l *Linter) checkShellInjection"):
        new_lines.extend(format_doc("Executes checkShellInjection operation.", ["None."], ["[]Result: A slice of findings."]))
        new_lines.append(line)
        i += 1
    elif line.startswith("func (l *Linter) checkInsecureHTTP"):
        new_lines.extend(format_doc("Executes checkInsecureHTTP operation.", ["None."], ["[]Result: A slice of findings."]))
        new_lines.append(line)
        i += 1
    elif line.startswith("func (l *Linter) checkCacheSettings"):
        new_lines.extend(format_doc("Executes checkCacheSettings operation.", ["None."], ["[]Result: A slice of findings."]))
        new_lines.append(line)
        i += 1
    elif line.startswith("//"): # Skip existing comments for these functions
        j = i
        while j < len(lines) and lines[j].startswith("//"): j += 1
        if j < len(lines) and (lines[j].startswith("func") or lines[j].startswith("type")):
            # It is a doc for a func or type, we will replace it or handle it
            # For simplicity, if we hit one of our targets, we consume the old doc
            if any(target in lines[j] for target in ["String()", "NewLinter", "Run(", "checkPlainText", "checkShell", "checkInsecure", "checkCache"]):
                i = j
                continue
        new_lines.append(line)
        i += 1
    else:
        new_lines.append(line)
        i += 1

with open(path, "w") as f:
    f.writelines(new_lines)
