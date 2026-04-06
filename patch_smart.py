import re

with open('ui/src/components/tools/smart-result-renderer.tsx', 'r') as f:
    content = f.read()

# Instead of hiding the tabs entirely and rendering separate buttons, let's just give these buttons the 'tab' role so the E2E tests work with them
content = re.sub(
    r'<Button\n(.*?)variant=\{activeView === "smart" \? "secondary" : "ghost"\}',
    r'<Button\n\1role="tab"\n                            aria-selected={activeView === "smart"}\n                            variant={activeView === "smart" ? "secondary" : "ghost"}',
    content
)

content = re.sub(
    r'<Button\n(.*?)variant=\{activeView === "rich" \? "secondary" : "ghost"\}',
    r'<Button\n\1role="tab"\n                            aria-selected={activeView === "rich"}\n                            variant={activeView === "rich" ? "secondary" : "ghost"}',
    content
)

content = re.sub(
    r'<Button\n(.*?)variant=\{activeView === "raw" \? "secondary" : "ghost"\}',
    r'<Button\n\1role="tab"\n                            aria-selected={activeView === "raw"}\n                            variant={activeView === "raw" ? "secondary" : "ghost"}',
    content
)

# And give the container role="tablist"
content = re.sub(
    r'<div className="flex items-center bg-muted/50 rounded-lg p-0.5 border">',
    r'<div className="flex items-center bg-muted/50 rounded-lg p-0.5 border" role="tablist">',
    content
)

with open('ui/src/components/tools/smart-result-renderer.tsx', 'w') as f:
    f.write(content)
