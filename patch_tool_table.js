const fs = require('fs');
const filePath = 'ui/src/components/tools/tool-table.tsx';
let content = fs.readFileSync(filePath, 'utf8');

content = content.replace(
  'usageStats?.[tool.name]?.calls',
  'usageStats?.[tool.name]?.call_count'
);

fs.writeFileSync(filePath, content);
