import re

with open('src/app/tools/page.tsx', 'r') as f:
    content = f.read()

# Make sure we import useMemo
if 'useMemo' not in content:
    content = content.replace('import { useState, useEffect } from "react";', 'import { useState, useEffect, useMemo } from "react";')

grouped_pattern = r'  // Grouping logic\n  const groupedTools = filteredTools\.reduce\(\(acc, tool\) => \{\n    let key = "Other";\n    if \(groupBy === "service"\) \{\n      const service = services\.find\(\(s\) => s\.id === tool\.serviceId\);\n      key = service \? service\.name : tool\.serviceId \|\| "Unknown Service";\n    \} else if \(groupBy === "category"\) \{\n      key = tool\.tags && tool\.tags\.length > 0 \? tool\.tags\[0\] : "Uncategorized";\n    \}\n\n    if \(!acc\[key\]\) \{\n      acc\[key\] = \[\];\n    \}\n    acc\[key\]\.push\(tool\);\n    return acc;\n  \}, \{\} as Record<string, ToolDefinition\[\]>\);'
replacement = '''  // ⚡ BOLT: [Render Optimization] Memoize grouping to prevent O(N^2) renders and redundant O(N) Array.find calls
  // Randomized Selection from Top 5 High-Impact Targets
  const groupedTools = useMemo(() => {
    const serviceMap = new Map(services.map(s => [s.id, s.name]));

    return filteredTools.reduce((acc, tool) => {
      let key = "Other";
      if (groupBy === "service") {
        const sName = serviceMap.get(tool.serviceId);
        key = sName ? sName : tool.serviceId || "Unknown Service";
      } else if (groupBy === "category") {
        key = tool.tags && tool.tags.length > 0 ? tool.tags[0] : "Uncategorized";
      }

      if (!acc[key]) {
        acc[key] = [];
      }
      acc[key].push(tool);
      return acc;
    }, {} as Record<string, ToolDefinition[]>);
  }, [filteredTools, services, groupBy]);'''

content = re.sub(grouped_pattern, replacement, content, flags=re.DOTALL)

with open('src/app/tools/page.tsx', 'w') as f:
    f.write(content)
