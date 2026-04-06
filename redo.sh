# 1. Move and update SmartResultRenderer -> RichResultViewer functionality
mv ui/src/components/playground/pro/smart-result-renderer.tsx ui/src/components/tools/
mv ui/src/components/playground/pro/smart-result-renderer.test.tsx ui/src/components/tools/

sed -i 's|../../tools/smart-table|./smart-table|g' ui/src/components/tools/smart-result-renderer.tsx
sed -i "s|import { SmartResultRenderer } from './smart-result-renderer';|import { SmartResultRenderer } from './smart-result-renderer';|g" ui/src/components/tools/smart-result-renderer.test.tsx
sed -i 's|@/components/playground/pro/smart-result-renderer|@/components/tools/smart-result-renderer|g' ui/src/tests/unit/smart-result-renderer.test.tsx
sed -i 's|import { SmartResultRenderer } from "./smart-result-renderer";|import { SmartResultRenderer } from "@/components/tools/smart-result-renderer";|g' ui/src/components/playground/pro/chat-message.tsx

cat << 'INNER_EOF' > ui/src/components/tools/rich-result-viewer.tsx
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { SmartResultRenderer } from "./smart-result-renderer";

interface RichResultViewerProps {
    result: any;
}

/**
 * Intent: Document RichResultViewer
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * RichResultViewer delegates to SmartResultRenderer for a unified rendering experience.
 *
 * @param props - The component props.
 * @param props.result - The raw result object from the tool execution.
 * @returns The rendered component.
 */
export function RichResultViewer({ result }: RichResultViewerProps) {
    return <SmartResultRenderer result={result} />;
}
INNER_EOF

cat << 'INNER_EOF' > ui/src/components/tools/rich-result-viewer.test.tsx
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { RichResultViewer } from './rich-result-viewer';

describe('RichResultViewer', () => {
    it('delegates to SmartResultRenderer and displays JSON mode initially', () => {
        const result = { test: 'value' };
        render(<RichResultViewer result={result} />);

        // SmartResultRenderer displays a JSON button when it renders raw JSON
        const jsonButtons = screen.getAllByRole('button', { name: /JSON/i });
        expect(jsonButtons.length).toBeGreaterThan(0);
    });
});
INNER_EOF

# 2. Update usage in Trace Detail & Payloads
sed -i 's|<JsonView data={trace.rootSpan.input} maxHeight={400} smartTable={true} />|<RichResultViewer result={trace.rootSpan.input} />|g' ui/src/components/traces/trace-detail.tsx
sed -i 's|<JsonView data={trace.rootSpan.output} maxHeight={400} smartTable={true} />|<RichResultViewer result={trace.rootSpan.output} />|g' ui/src/components/traces/trace-detail.tsx
sed -i 's|<JsonView data={safeParsePayload(trace.rootSpan.input)} maxHeight={400} smartTable={true} />|<RichResultViewer result={safeParsePayload(trace.rootSpan.input)} />|g' ui/src/components/dashboard/recent-activity-widget.tsx
sed -i 's|<JsonView data={safeParsePayload(trace.rootSpan.output)} maxHeight={400} smartTable={true} />|<RichResultViewer result={safeParsePayload(trace.rootSpan.output)} />|g' ui/src/components/dashboard/recent-activity-widget.tsx
sed -i 's|<JsonView data={data} smartTable={true} />|<RichResultViewer result={data} />|g' ui/src/components/logs/json-viewer.tsx
sed -i 's|<JsonView data={jsonContent} smartTable={true} maxHeight={400} />|<RichResultViewer result={jsonContent} />|g' ui/src/components/logs/log-viewer.tsx

sed -i 's|import { JsonView } from "@/components/ui/json-view";|import { JsonView } from "@/components/ui/json-view";\nimport { RichResultViewer } from "@/components/tools/rich-result-viewer";|g' ui/src/components/dashboard/recent-activity-widget.tsx
sed -i 's|import { JsonView } from "@/components/ui/json-view";|import { JsonView } from "@/components/ui/json-view";\nimport { RichResultViewer } from "@/components/tools/rich-result-viewer";|g' ui/src/components/logs/json-viewer.tsx
sed -i 's|const JsonView = lazy(() => import('\''@/components/ui/json-view'\'').then(m => ({ default: m.JsonView })));|const JsonView = lazy(() => import('\''@/components/ui/json-view'\'').then(m => ({ default: m.JsonView })));\nimport { RichResultViewer } from "@/components/tools/rich-result-viewer";|g' ui/src/components/logs/log-viewer.tsx

# 3. Clean up JsonView
sed -i '/import { SmartTable }/d' ui/src/components/ui/json-view.tsx

cat << 'INNER_EOF' > patch_json_view.py
import re

with open('ui/src/components/ui/json-view.tsx', 'r') as f:
    content = f.read()

content = re.sub(r'const getTableData = \(data: unknown, smartTable: boolean\) => \{.*?\};\n\n', '', content, flags=re.DOTALL)
content = re.sub(r'\s*smartTable\?: boolean;', '', content)
content = re.sub(r'\s*\* @param props.smartTable - Whether to attempt smart table rendering.', '', content)
content = re.sub(r'smartTable = true, ', '', content)
content = re.sub(r'\s*const tableData = useMemo\(\(\) => getTableData\(data, smartTable\), \[data, smartTable\]\);', '', content)
content = re.sub(r'const hasSmartView = tableData !== null;', 'const hasSmartView = false;', content)
content = re.sub(r'const hasSmartView = false;\n\s*const hasSmartView = false;', 'const hasSmartView = false;', content)
content = re.sub(r', Table as TableIcon', '', content)
content = re.sub(r'const renderSmart = \(\) => \{.*?\};\n\n', '', content, flags=re.DOTALL)
content = re.sub(r'const tableData = getTableData\(data, smartTable\);\n\s*if \(tableData\) return "smart";', '', content)

with open('ui/src/components/ui/json-view.tsx', 'w') as f:
    f.write(content)
INNER_EOF
python3 patch_json_view.py

sed -i 's| smartTable={false}||g' ui/src/components/tools/smart-result-renderer.tsx
sed -i 's| smartTable={true}||g' ui/src/components/ui/json-view.test.tsx
sed -i '/supports smart table view/,$d' ui/src/components/ui/json-view.test.tsx

cat << 'INNER_EOF' >> ui/src/components/ui/json-view.test.tsx
  it('collapses long content', () => {
      const data = { key: 'very long content' };
      render(<JsonView data={data} maxHeight={100} />);
      expect(screen.getByText('Show More')).toBeInTheDocument();
      fireEvent.click(screen.getByText('Show More'));
      expect(screen.getByText('Show Less')).toBeInTheDocument();
  });
});
INNER_EOF
