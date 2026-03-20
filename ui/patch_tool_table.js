const fs = require('fs');
const filePath = 'src/components/tools/tool-table.tsx';
let content = fs.readFileSync(filePath, 'utf8');

content = content.replace(
  'usageStats?.[tool.name]?.call_count',
  'usageStats?.[tool.name]?.totalCalls'
);

fs.writeFileSync(filePath, content);
