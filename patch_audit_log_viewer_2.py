import sys
import re

def modify_file():
    filepath = "ui/src/components/audit/audit-log-viewer.tsx"
    with open(filepath, 'r') as f:
        content = f.read()

    # The current JSON.parse will fail if selectedLog.arguments is not valid JSON
    # It's better to try to parse or fallback to string.
    # RichResultViewer handles both object and string automatically, so we just need to pass the parsed JSON if possible
    # We can write a quick utility inside the component:

    safe_parse = """
    const safeParse = (str: string | undefined | null) => {
        if (!str) return null;
        try {
            return JSON.parse(str);
        } catch (e) {
            return str;
        }
    };
"""

    if "safeParse" not in content:
        # insert before return (
        idx = content.find("return (")
        content = content[:idx] + safe_parse + "\n    " + content[idx:]

    args_replacement = '<RichResultViewer result={safeParse(selectedLog.arguments) || {}} />'
    result_replacement = '<RichResultViewer result={safeParse(selectedLog.result) || (selectedLog.error ? null : {})} />'

    content = re.sub(r'<RichResultViewer result=\{selectedLog\.arguments \? JSON\.parse\(selectedLog\.arguments\) : \{\}\} />', args_replacement, content)
    content = re.sub(r'<RichResultViewer result=\{selectedLog\.result \? JSON\.parse\(selectedLog\.result\) : \(selectedLog\.error \? null : \{\}\)\} />', result_replacement, content)

    with open(filepath, 'w') as f:
        f.write(content)

if __name__ == "__main__":
    modify_file()
