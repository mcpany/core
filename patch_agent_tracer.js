const fs = require('fs');
let content = fs.readFileSync('ui/src/components/dashboard/agent-chain-tracer.tsx', 'utf8');

// There are duplicate chainData variables!
content = content.replace(/  const { traces } = useTraces\(\{ limit: 10 \}\);\n\n  const chainData = traces\.slice\(0, 5\)\.map\(\(trace\) => \{\n[\s\S]*?  \}\);\n\n  const \[expandedId, setExpandedId\] = useState<string \| null>\(null\);\n  const { traces } = useTraces\(\);/g, `  const [expandedId, setExpandedId] = useState<string | null>(null);
  const { traces } = useTraces();`);

fs.writeFileSync('ui/src/components/dashboard/agent-chain-tracer.tsx', content);
