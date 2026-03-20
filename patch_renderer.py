import re

with open('ui/src/components/playground/pro/smart-result-renderer.tsx', 'r') as f:
    content = f.read()

# Replace renderRaw to gracefully handle text if it's a string, falling back to JSON view if it's an object
old_render = """    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );"""

new_render = """    const renderRaw = () => {
        if (typeof result === 'string') {
            return (
                <div className="whitespace-pre-wrap font-mono text-sm p-3 rounded-md border bg-muted/10 max-h-[400px] overflow-auto">
                    {result}
                </div>
            );
        }
        return <JsonView data={result} maxHeight={400} />;
    };"""

content = content.replace(old_render, new_render)

with open('ui/src/components/playground/pro/smart-result-renderer.tsx', 'w') as f:
    f.write(content)
