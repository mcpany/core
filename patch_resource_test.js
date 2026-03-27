const fs = require('fs');

let content = fs.readFileSync('ui/src/components/resources/resource-preview-modal.test.tsx', 'utf8');

// The original mock
// const MockHighlighter = ({ children }: { children: React.ReactNode }) => <pre data-testid="code-block">{children}</pre>;

// Update it to properly handle the way it's used in the component
content = content.replace(
  "const MockHighlighter = ({ children }: { children: React.ReactNode }) => <pre data-testid=\"code-block\">{children}</pre>;",
  "const MockHighlighter = ({ children, value }: { children?: React.ReactNode, value?: string }) => <pre data-testid=\"code-block\">{value || children}</pre>;"
);

fs.writeFileSync('ui/src/components/resources/resource-preview-modal.test.tsx', content);
