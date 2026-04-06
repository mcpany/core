import re

with open('ui/src/components/ui/json-view.tsx', 'r') as f:
    content = f.read()

content = re.sub(r'const getTableData = \(data: unknown, smartTable: boolean\) => \{.*?\};\n\n', '', content, flags=re.DOTALL)
content = re.sub(r'\s*smartTable\?: boolean;', '', content)
content = re.sub(r'\s*\* @param props.smartTable - Whether to attempt smart table rendering.', '', content)
content = re.sub(r'smartTable = true, ', '', content)
content = re.sub(r'\s*const tableData = useMemo\(\(\) => getTableData\(data, smartTable\), \[data, smartTable\]\);', '', content)
content = re.sub(r'const hasSmartView = tableData !== null;', 'const hasSmartView = false;', content)
content = re.sub(r'const hasSmartView = false;\n\s*const hasSmartView = false;', 'const hasSmartView = false;', content)
content = re.sub(r', Table as TableIcon', '', content)
content = re.sub(r'const renderSmart = \(\) => \{.*?\};\n\n', '', content, flags=re.DOTALL)
content = re.sub(r'const tableData = getTableData\(data, smartTable\);\n\s*if \(tableData\) return "smart";', '', content)

with open('ui/src/components/ui/json-view.tsx', 'w') as f:
    f.write(content)
