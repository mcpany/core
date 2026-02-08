/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, Image as ImageIcon, FileText } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";

/**
 * Props for the SmartResultRenderer component.
 */
interface SmartResultRendererProps {
    /** The result object to render. Can be a JSON string, an object, or an array. */
    result: any;
}

interface McpContent {
  type: string;
  text?: string;
  data?: string;
  mimeType?: string;
  resource?: any;
}

function isMcpContent(data: any): data is McpContent[] {
  if (!Array.isArray(data) || data.length === 0) return false;
  return data.every(item =>
    typeof item === 'object' && item !== null &&
    (item.type === 'text' || item.type === 'image' || item.type === 'resource')
  );
}

function RichContentRenderer({ content }: { content: McpContent[] }) {
    return (
      <div className="flex flex-col gap-4 p-4 bg-muted/10 rounded-md border text-sm">
        {content.map((item, i) => (
          <div key={i} className="w-full overflow-auto">
            {item.type === 'image' && item.data && (
               <div className="flex flex-col gap-2 items-start">
                  <img
                    src={`data:${item.mimeType || 'image/png'};base64,${item.data}`}
                    alt={`Result Image ${i+1}`}
                    className="max-w-full h-auto rounded border bg-[url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAMUlEQVQ4T2NkYGAQYcAP3uCTZhw1gGGYhAGBZIA/nYDCgBDAm9BGDWAAjyQc6wcE0AwAv16BC1L/4hcmAAAAAElFTkSuQmCC')] bg-repeat"
                  />
                  <div className="text-xs text-muted-foreground font-mono flex items-center gap-1">
                      <ImageIcon className="h-3 w-3" />
                      {item.mimeType || 'image/png'}
                  </div>
               </div>
            )}
            {item.type === 'text' && item.text && (
               <div className="flex flex-col gap-1">
                   {content.length > 1 && (
                        <div className="text-xs text-muted-foreground font-mono flex items-center gap-1 mb-1">
                            <FileText className="h-3 w-3" />
                            Text Output
                        </div>
                   )}
                   <pre className="whitespace-pre-wrap font-mono text-xs text-foreground/90 bg-muted/30 p-2 rounded">
                      {item.text}
                   </pre>
               </div>
            )}
            {item.type === 'resource' && (
                <div className="p-2 border rounded bg-muted/20">
                    <pre className="text-xs">{JSON.stringify(item.resource, null, 2)}</pre>
                </div>
            )}
          </div>
        ))}
      </div>
    )
  }

/**
 * Renders the result of a tool execution in a smart, tabular format if possible,
 * falling back to a raw JSON view.
 *
 * @param props - The component props.
 * @param props.result - The result object to render. Can be a JSON string, an object, or an array.
 * @returns A React component that displays the result.
 */
export function SmartResultRenderer({ result }: SmartResultRendererProps) {
    const [viewMode, setViewMode] = useState<"smart" | "raw">("smart");

    // Attempt to parse a tabular structure from the result
    const tableData = useMemo(() => {
        let content = result;

        // 1. Unwrap CallToolResult structure
        if (result && typeof result === 'object' && Array.isArray(result.content)) {
            // Check if it is MCP content array
            if (isMcpContent(result.content)) {
                 // If it has images, or multiple items, we treat the content array as the data
                 // If it is single text item, we try to unwrap it to parse JSON inside it
                 const hasImage = result.content.some((c: any) => c.type === 'image');
                 const hasMultiple = result.content.length > 1;

                 if (hasImage || hasMultiple) {
                     content = result.content;
                 } else {
                     // Single text item, try to extract text to parse as JSON table
                     const textItem = result.content.find((c: any) => c.type === 'text');
                     if (textItem) {
                         content = textItem.text;
                     } else {
                         content = result.content;
                     }
                 }
            } else {
                 const textItem = result.content.find((c: any) => c.type === 'text' || (c.text && !c.type));
                 if (textItem) {
                     content = textItem.text;
                 }
            }
        }

        // 2. Parse JSON string if it's a string
        if (typeof content === 'string') {
            try {
                content = JSON.parse(content);
            } catch (e) {
                // Not valid JSON string
                // If content was a string (e.g. from single text item), keep it as is?
                // But downstream expects array of objects for table.
                // If parsing fails, content remains string.
                // Step 4 will fail. tableData = null. Raw view will show string. Correct.
            }
        }

        // 3. Handle Command Output wrapper (e.g. { stdout: "[...]", ... })
        if (content && typeof content === 'object' && !Array.isArray(content)) {
             if (content.stdout && typeof content.stdout === 'string') {
                 try {
                     const inner = JSON.parse(content.stdout);
                     // If inner is array (maybe MCP content or data array)
                     if (Array.isArray(inner)) {
                         // Check if inner is nested MCP content output by a command
                         if (inner.length > 0 && inner[0].content && Array.isArray(inner[0].content)) {
                             // This handles case where command outputs CallToolResult JSON
                             // We don't support double nesting unwrapping here yet easily without recursion
                             // But wait, if command outputs `{"content": [...]}`, inner is Object, not Array.
                             // So the checks above handle "inner is Array".
                             // If command outputs `[{"type": "image", ...}]`, inner is Array.
                             content = inner;
                         } else if (typeof inner === 'object' && inner.content && Array.isArray(inner.content)) {
                             // Command output is a JSON object that mimics CallToolResult
                             content = inner.content;
                         } else {
                             content = inner;
                         }
                     } else if (typeof inner === 'object' && inner !== null) {
                         // If command outputs `{"content": ...}` (CallToolResult)
                         if (inner.content && Array.isArray(inner.content)) {
                             content = inner.content;
                         } else {
                             content = inner;
                         }
                     }
                 } catch (e) {
                     // stdout is not JSON
                 }
             } else if (content.content && Array.isArray(content.content)) {
                  // Maybe nested content?
                  content = content.content;
             }
        }

        // 4. Final validation: Must be an array of objects
        if (Array.isArray(content) && content.length > 0) {
            // Check if items are objects
            const isListOfObjects = content.every((item: any) => typeof item === 'object' && item !== null && !Array.isArray(item));
            if (isListOfObjects) {
                return content;
            }
        }

        return null;
    }, [result]);

    const mcpContent = useMemo(() => {
        if (tableData && isMcpContent(tableData)) {
             // If ANY item is image, or multiple items, prefer Rich View
             const hasImage = tableData.some(c => c.type === 'image');

             // If tableData is MCP content, it has 'type', 'text'/'data'.
             // If we just show it as a Table, it shows columns "type", "text".
             // If type is 'text' and text is JSON string, Table view is UGLY (shows escaped JSON).
             // If type is 'text' and text is simple string, Table view is ok but redundant ("text": "value").

             // Generally, if it matches McpContent structure, we probably want Rich view
             // UNLESS it was explicitly parsed from a JSON array intended to be data.

             // But if `isMcpContent` is true, it means it HAS `type: 'text'|'image'`.
             // Normal data usually doesn't strictly follow this unless it IS McpContent.
             // So safe to assume we want Rich View.

             return tableData as McpContent[];
        }
        return null;
    }, [tableData]);


    const hasSmartView = tableData !== null;
    const showRichContent = viewMode === "smart" && mcpContent !== null;

    const renderRaw = () => (
        <JsonView data={result} maxHeight={400} />
    );

    const renderSmart = () => {
        if (!tableData) return renderRaw();

        // Determine columns from all keys in the first 10 rows (to be reasonably safe)
        const allKeys = new Set<string>();
        tableData.slice(0, 10).forEach((row: any) => {
            Object.keys(row).forEach(k => allKeys.add(k));
        });
        const columns = Array.from(allKeys);

        return (
            <div className="rounded-md border max-h-[400px] overflow-auto">
                <Table>
                    <TableHeader className="bg-muted/50 sticky top-0">
                        <TableRow>
                            {columns.map(col => (
                                <TableHead key={col} className="whitespace-nowrap font-medium text-xs px-2 py-1 h-8">
                                    {col}
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {tableData.map((row: any, idx: number) => (
                            <TableRow key={idx} className="hover:bg-muted/50">
                                {columns.map(col => {
                                    const val = row[col];
                                    let displayVal = val;
                                    if (typeof val === 'object' && val !== null) {
                                        displayVal = JSON.stringify(val);
                                    } else if (typeof val === 'boolean') {
                                        displayVal = val ? "true" : "false";
                                    }

                                    return (
                                        <TableCell key={col} className="px-2 py-1 text-xs max-w-[200px] truncate" title={String(displayVal)}>
                                            {String(displayVal ?? "")}
                                        </TableCell>
                                    );
                                })}
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
                <div className="bg-muted/30 px-2 py-1 text-[10px] text-muted-foreground border-t">
                    Showing {tableData.length} rows
                </div>
            </div>
        );
    };

    return (
        <div className="flex flex-col gap-0 w-full">
            {hasSmartView && (
                <div className="flex justify-end mb-1 px-1">
                     <div className="flex items-center bg-muted/50 rounded-lg p-0.5 border">
                         <Button
                            variant={viewMode === "smart" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("smart")}
                         >
                             {showRichContent ? <ImageIcon className="size-3" /> : <TableIcon className="size-3" />}
                             {showRichContent ? "Visual" : "Table"}
                         </Button>
                         <Button
                            variant={viewMode === "raw" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("raw")}
                         >
                             <Code className="size-3" /> JSON
                         </Button>
                     </div>
                </div>
            )}

            {showRichContent ? <RichContentRenderer content={mcpContent!} /> : (viewMode === "smart" && hasSmartView ? renderSmart() : renderRaw())}
        </div>
    );
}
