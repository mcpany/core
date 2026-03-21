/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, Image as ImageIcon, FileText } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";
import { unwrapMcpResult } from "@/lib/mcp-unwrap";
import { DataTable } from "@/components/ui/data-table";

/**
 * Props for the SmartResultRenderer component.
 */
interface SmartResultRendererProps {
    /** The result object to render. Can be a JSON string, an object, or an array. */
    result: unknown;
}

interface McpContent {
    type: 'text' | 'image' | 'resource';
    text?: string;
    data?: string;
    mimeType?: string;
    resource?: Record<string, unknown>;
}

/**
 * Renders the result of a tool execution in a smart, tabular format if possible,
 * falling back to a raw JSON view.
 */
export function SmartResultRenderer({ result }: SmartResultRendererProps) {
    const [userViewMode, setUserViewMode] = useState<"smart" | "raw" | "rich" | null>(null);

    // 1. Shared unwrapping logic
    const unwrappedContent = useMemo(() => {
        // First unwrap the main wrapper without parsing inner text yet
        let content = result;

        // Unwrap CallToolResult structure
        if (result && typeof result === 'object' && Array.isArray((result as Record<string, unknown>).content)) {
            content = (result as Record<string, unknown>).content;
        }

        // Handle Command Output wrapper
        if (content && typeof content === 'object' && !Array.isArray(content)) {
             if ((content as Record<string, unknown>).stdout && typeof (content as Record<string, unknown>).stdout === 'string') {
                 try {
                     const inner = JSON.parse((content as Record<string, unknown>).stdout as string);
                     if (Array.isArray(inner) || (typeof inner === 'object' && inner !== null)) {
                         content = inner;
                     }
                 } catch (_) {
                     // stdout is not JSON
                 }
             }
        }

        // Handle deeply nested "content" (e.g. from stdout containing MCP content object)
        if (content && typeof content === 'object' && !Array.isArray(content) && Array.isArray((content as Record<string, unknown>).content)) {
            content = (content as Record<string, unknown>).content;
        }

        return content;
    }, [result]);

    const fullyUnwrapped = useMemo(() => unwrapMcpResult(result), [result]);

    // 2. Identify MCP Content
    const mcpContent = useMemo<McpContent[] | null>(() => {
        if (Array.isArray(unwrappedContent) && unwrappedContent.length > 0) {
            const isMcp = unwrappedContent.every((item: unknown) =>
                typeof item === 'object' && item !== null &&
                ((item as Record<string, unknown>).type === 'text' || (item as Record<string, unknown>).type === 'image' || (item as Record<string, unknown>).type === 'resource')
            );
            if (isMcp) return unwrappedContent as McpContent[];
        }
        return null;
    }, [unwrappedContent]);

    // 3. Identify Table Data
    const tableData = useMemo(() => {
        // If fullyUnwrapped is a table, and there's no MCP content with non-text elements, use it directly!
        if (Array.isArray(fullyUnwrapped) && fullyUnwrapped.length > 0) {
             const isTable = fullyUnwrapped.every((item: unknown) => typeof item === 'object' && item !== null);
             // Make sure we don't accidentally treat MCP image objects as a table row if we want rich view
             const isRichMcp = mcpContent && mcpContent.some(c => c.type !== 'text');
             if (isTable && !isRichMcp) return fullyUnwrapped;
        }

        // If MCP content, try to extract table data from text
        if (mcpContent) {
             const hasNonText = mcpContent.some(c => c.type !== 'text');
             if (hasNonText) return null;

             // Only support single text block for table view to avoid complexity
             if (mcpContent.length === 1 && mcpContent[0].text) {
                 try {
                    const parsed = JSON.parse(mcpContent[0].text);
                    if (Array.isArray(parsed) && parsed.every(item => typeof item === 'object')) {
                        return parsed;
                    }
                } catch (_) {}
             }
             return null;
        }

        // If NOT MCP content, check if unwrapped content itself is tabular data (CLI use case)
        if (Array.isArray(unwrappedContent) && unwrappedContent.length > 0) {
             const isTable = unwrappedContent.every((item: unknown) => typeof item === 'object' && item !== null);
             if (isTable) return unwrappedContent as Record<string, unknown>[];
        }

        return null;
    }, [fullyUnwrapped, unwrappedContent, mcpContent]);

    const activeView = useMemo(() => {
        // User override
        if (userViewMode === 'smart' && tableData) return 'smart';
        if (userViewMode === 'rich' && mcpContent) return 'rich';
        if (userViewMode === 'raw') return 'raw';

        // Auto defaults (if user mode invalid or null)
        if (tableData) return 'smart';
        if (mcpContent) return 'rich';
        return 'raw';
    }, [userViewMode, tableData, mcpContent]);

    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );

    const renderRich = () => {
        if (!mcpContent) return renderRaw();

        return (
            <div className="flex flex-col gap-4 p-4 border rounded-md bg-muted/10">
                {mcpContent.map((item, idx) => (
                    <div key={idx} className="flex flex-col gap-2">
                        {item.type === 'text' && (
                            <div className="whitespace-pre-wrap font-mono text-sm bg-muted/30 p-3 rounded-md border border-white/5">
                                {item.text}
                            </div>
                        )}
                        {item.type === 'image' && item.data && (
                            <div className="flex flex-col gap-1 items-start">
                                <img
                                    src={`data:${item.mimeType || 'image/png'};base64,${item.data}`}
                                    alt="Tool Result"
                                    className="max-w-full h-auto rounded-lg border border-white/10 shadow-sm"
                                />
                                <span className="text-[10px] text-muted-foreground self-end">
                                    {item.mimeType}
                                </span>
                            </div>
                        )}
                        {item.type === 'resource' && (
                            <div className="flex items-center gap-2 p-3 bg-muted/30 rounded-md border border-white/5">
                                <FileText className="h-4 w-4 text-primary" />
                                <span className="text-sm font-medium">Resource: {String(item.resource?.uri || 'Unknown')}</span>
                            </div>
                        )}
                    </div>
                ))}
            </div>
        );
    };

    const renderSmartTable = () => {
        if (!tableData) return null;

        // Determine columns from all keys in the first 10 rows
        const allKeys = new Set<string>();
        tableData.slice(0, 10).forEach((row: Record<string, unknown>) => {
            Object.keys(row).forEach(k => allKeys.add(k));
        });
        const columns = Array.from(allKeys).map(col => ({
            accessorKey: col,
            header: col,
            cell: (row: Record<string, unknown>) => {
                const val = row[col];
                let displayVal = val;
                if (typeof val === 'object' && val !== null) {
                    displayVal = JSON.stringify(val);
                } else if (typeof val === 'boolean') {
                    displayVal = val ? "true" : "false";
                }
                return String(displayVal ?? "");
            }
        }));

        return (
             <div className="w-full pt-2">
                 <DataTable data={tableData as Record<string, unknown>[]} columns={columns} />
             </div>
        );
    };

    return (
        <div className="flex flex-col gap-0 w-full">
            <div className="flex justify-end mb-1 px-1">
                 <div className="flex items-center bg-muted/50 rounded-lg p-0.5 border">
                     {tableData && (
                        <Button
                            variant={activeView === "smart" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1 transition-colors"
                            onClick={() => setUserViewMode("smart")}
                        >
                            <TableIcon className="size-3" /> Table
                        </Button>
                     )}
                     {mcpContent && (
                        <Button
                            variant={activeView === "rich" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1 transition-colors"
                            onClick={() => setUserViewMode("rich")}
                        >
                            <ImageIcon className="size-3" /> Rich
                        </Button>
                     )}
                     <Button
                        variant={activeView === "raw" ? "secondary" : "ghost"}
                        size="sm"
                        className="h-6 px-2 text-[10px] gap-1 transition-colors"
                        onClick={() => setUserViewMode("raw")}
                     >
                         <Code className="size-3" /> JSON
                     </Button>
                 </div>
            </div>

            {activeView === 'smart' && renderSmartTable()}
            {activeView === 'rich' && renderRich()}
            {activeView === 'raw' && renderRaw()}
        </div>
    );
}
