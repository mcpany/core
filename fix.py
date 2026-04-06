import re

with open('ui/src/components/tools/rich-result-viewer.tsx', 'r') as f:
    content = f.read()

# The raw tab should explicitly disable smartTable because it's meant to be raw
new_content = content.replace('<TabsContent value="raw">\n                    <JsonView data={result} maxHeight={400} smartTable={true} />\n                </TabsContent>', '<TabsContent value="raw">\n                    <JsonView data={result} maxHeight={400} smartTable={false} />\n                </TabsContent>')

with open('ui/src/components/tools/rich-result-viewer.tsx', 'w') as f:
    f.write(new_content)
