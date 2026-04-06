import re

with open('ui/src/components/ui/json-view.tsx', 'r') as f:
    content = f.read()

# Remove the Table button since we don't use hasSmartView anymore
content = re.sub(r'\{hasSmartView && \([\s\S]*?\{viewMode === "smart" \? "secondary" : "ghost"\}[\s\S]*?<TableIcon className="size-3" /> Table\n\s*</Button>\n\s*\)\}', '', content)

# Update the render logic
content = re.sub(r'\{viewMode === "smart" && hasSmartView \? renderSmart\(\) :\s*\n', '{', content)

with open('ui/src/components/ui/json-view.tsx', 'w') as f:
    f.write(content)
