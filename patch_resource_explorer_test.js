const fs = require('fs');

let files = [
  'ui/src/components/resources/resource-explorer.test.tsx',
  'ui/src/components/resources/resource-viewer.test.tsx'
];

for (let file of files) {
  if (fs.existsSync(file)) {
    let content = fs.readFileSync(file, 'utf8');

    // Apply the same fix to any other files that might be mocking the highlighter incorrectly
    content = content.replace(
      /const MockHighlighter = \(\{ children \}: \{ children: React\.ReactNode \}\) => <pre data-testid="code-block">\{children\}<\/pre>;/g,
      "const MockHighlighter = ({ children, value }: { children?: React.ReactNode, value?: string }) => <pre data-testid=\"code-block\">{value || children}</pre>;"
    );

    fs.writeFileSync(file, content);
  }
}
