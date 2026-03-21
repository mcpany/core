import sys

def modify_file():
    filepath = "ui/src/components/audit/audit-log-viewer.tsx"
    with open(filepath, 'r') as f:
        content = f.read()

    # Import RichResultViewer if not present
    if "RichResultViewer" not in content:
        import_statement = "import { RichResultViewer } from \"@/components/tools/rich-result-viewer\";\n"
        # Find the last import statement
        last_import_idx = content.rfind("import ")
        end_of_line_idx = content.find("\n", last_import_idx)
        content = content[:end_of_line_idx + 1] + import_statement + content[end_of_line_idx + 1:]

    # Remove SyntaxHighlighter usage and use RichResultViewer instead
    # Search for `<SyntaxHighlighter ...> {formatJson(selectedLog.arguments) || "{}"} </SyntaxHighlighter>`

    # We will replace the entire blocks for Arguments and Result
    import re

    args_block_pattern = re.compile(r'<h4 className="text-sm font-medium mb-2">Arguments</h4>\s*<div className="rounded-md overflow-hidden border">\s*<SyntaxHighlighter[^>]*>\s*\{formatJson\(selectedLog\.arguments\) \|\| "\{\}"\}\s*</SyntaxHighlighter>\s*</div>')
    result_block_pattern = re.compile(r'<h4 className="text-sm font-medium mb-2">Result</h4>\s*<div className="rounded-md overflow-hidden border">\s*<SyntaxHighlighter[^>]*>\s*\{formatJson\(selectedLog\.result\) \|\| \(selectedLog\.error \? "null" : "\{\}"\)\}\s*</SyntaxHighlighter>\s*</div>')

    new_args_block = '<h4 className="text-sm font-medium mb-2">Arguments</h4>\n                                <div className="rounded-md overflow-hidden border">\n                                    <RichResultViewer result={selectedLog.arguments ? JSON.parse(selectedLog.arguments) : {}} />\n                                </div>'
    new_result_block = '<h4 className="text-sm font-medium mb-2">Result</h4>\n                                <div className="rounded-md overflow-hidden border">\n                                    <RichResultViewer result={selectedLog.result ? JSON.parse(selectedLog.result) : (selectedLog.error ? null : {})} />\n                                </div>'

    content = args_block_pattern.sub(new_args_block, content)
    content = result_block_pattern.sub(new_result_block, content)

    # Need to remove the formatJson function if it's unused
    format_json_pattern = re.compile(r'    const formatJson = \(jsonStr: string\) => \{[^}]*\} catch \(e\) \{[^}]*\}\s*\};\s*')
    content = format_json_pattern.sub('', content)

    # Remove SyntaxHighlighter imports if they are unused
    if "SyntaxHighlighter" not in content.replace("import SyntaxHighlighter", ""):
        content = re.sub(r'import SyntaxHighlighter from \'react-syntax-highlighter/dist/esm/light\';\n', '', content)
        content = re.sub(r'import json from \'react-syntax-highlighter/dist/esm/languages/hljs/json\';\n', '', content)
        content = re.sub(r'import vs2015 from \'react-syntax-highlighter/dist/esm/styles/hljs/vs2015\';\n', '', content)
        content = re.sub(r'SyntaxHighlighter\.registerLanguage\(\'json\', json\);\n', '', content)

    with open(filepath, 'w') as f:
        f.write(content)

if __name__ == "__main__":
    modify_file()
