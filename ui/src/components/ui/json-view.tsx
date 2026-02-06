/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, forwardRef } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, Copy, Check, ChevronDown, ChevronUp } from "lucide-react";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/use-toast";
import { TableVirtuoso, TableVirtuosoProps } from "react-virtuoso";

interface JsonViewProps {
  data: unknown;
  className?: string;
  /**
   * If true, attempts to render array of objects as a table.
   */
  smartTable?: boolean;
  /**
   * Max height in pixels before collapsing. Default: 400.
   * Set to 0 or negative to disable collapsing.
   */
  maxHeight?: number;
}

// ⚡ BOLT: Custom Table components for Virtuoso to match Shadcn UI styling.
// Randomized Selection from Top 5 High-Impact Targets (Virtualization)
const VirtuosoTableComponents: TableVirtuosoProps<any, any>['components'] = {
  Table: ({ style, ...props }: any) => (
    <table {...props} style={{ ...style, width: "100%" }} className="w-full caption-bottom text-sm" />
  ),
  TableBody: forwardRef(({ className, ...props }: any, ref) => (
    <tbody ref={ref} className={cn("[&_tr:last-child]:border-0", className)} {...props} />
  )),
  TableHead: forwardRef(({ className, ...props }: any, ref) => (
    <thead ref={ref} className={cn("[&_tr]:border-b bg-muted/50", className)} {...props} />
  )),
  TableRow: ({ className, ...props }: any) => (
    <tr className={cn("border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted", className)} {...props} />
  ),
};

/**
 * JsonView component.
 * Renders data with syntax highlighting, optional smart table view, and copy functionality.
 *
 * @param props - The component props.
 * @param props.data - The data to display.
 * @param props.className - The className.
 * @param props.smartTable - Whether to attempt smart table rendering.
 * @param props.maxHeight - Max height before collapsing.
 * @returns The rendered component.
 */
export function JsonView({ data, className, smartTable = false, maxHeight = 400 }: JsonViewProps) {
  const [viewMode, setViewMode] = useState<"smart" | "raw">("smart");
  const [copied, setCopied] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const { toast } = useToast();

  // Attempt to parse a tabular structure from the result if smartTable is enabled
  const tableData = useMemo(() => {
    if (!smartTable) return null;

    let content = data;

    // If it's a string, try to parse it
    if (typeof content === 'string') {
        try {
            content = JSON.parse(content);
        } catch (_e) {
            return null;
        }
    }

    // Must be an array of objects
    if (Array.isArray(content) && content.length > 0) {
        const isListOfObjects = content.every(item => typeof item === 'object' && item !== null && !Array.isArray(item));
        if (isListOfObjects) {
            return content;
        }
    }

    return null;
  }, [data, smartTable]);

  const handleCopy = () => {
    const text = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).catch(e => console.error("Clipboard error", e));
    }
    setCopied(true);
    toast({
        title: "Copied",
        description: "JSON copied to clipboard",
    });
    setTimeout(() => setCopied(false), 2000);
  };

  const hasSmartView = tableData !== null;
  const showCollapse = maxHeight > 0;

  // Calculate approximate lines to guess if we need expand button without rendering?
  // Hard to do accurately. We'll use CSS max-height.

  const renderRaw = () => (
    <div className={cn("relative group/code rounded-md bg-[#1e1e1e]", className)}>
        <div
            className={cn(
                "overflow-hidden transition-all",
                showCollapse && !isExpanded ? "relative" : ""
            )}
            style={{
                maxHeight: showCollapse && !isExpanded ? `${maxHeight}px` : undefined
            }}
        >
            <SyntaxHighlighter
                language="json"
                style={vscDarkPlus}
                customStyle={{
                    margin: 0,
                    padding: '1rem',
                    borderRadius: '0.375rem',
                    fontSize: '12px',
                    lineHeight: '1.5',
                    backgroundColor: 'transparent' // We set bg on parent
                }}
                wrapLines={true}
                wrapLongLines={true}
            >
                {typeof data === 'string' ? data : JSON.stringify(data, null, 2)}
            </SyntaxHighlighter>

            {showCollapse && !isExpanded && (
                <div className="absolute bottom-0 left-0 right-0 h-12 bg-gradient-to-t from-[#1e1e1e] to-transparent pointer-events-none" />
            )}
        </div>

        <div className="absolute right-2 top-2 flex gap-1 opacity-0 group-hover/code:opacity-100 transition-opacity">
             <Button
                size="icon"
                variant="ghost"
                className="h-6 w-6 bg-white/10 hover:bg-white/20 text-white"
                onClick={handleCopy}
                title="Copy JSON"
            >
                {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
            </Button>
        </div>

        {showCollapse && (
            <div className="flex justify-center p-1 border-t border-white/10 bg-[#1e1e1e] rounded-b-md">
                <Button
                    variant="ghost"
                    size="sm"
                    className="h-5 text-[10px] text-muted-foreground hover:text-white w-full"
                    onClick={() => setIsExpanded(!isExpanded)}
                >
                    {isExpanded ? (
                        <span className="flex items-center gap-1"><ChevronUp className="h-3 w-3" /> Show Less</span>
                    ) : (
                         <span className="flex items-center gap-1"><ChevronDown className="h-3 w-3" /> Show More</span>
                    )}
                </Button>
            </div>
        )}
    </div>
  );

  const renderSmart = () => {
    if (!tableData) return renderRaw();

    // Determine columns from all keys in the first 10 rows
    const allKeys = new Set<string>();
    tableData.slice(0, 10).forEach((row: Record<string, unknown>) => {
        Object.keys(row).forEach(k => allKeys.add(k));
    });
    const columns = Array.from(allKeys);

    // ⚡ BOLT: Implemented virtualization for JSON table view using react-virtuoso.
    // Randomized Selection from Top 5 High-Impact Targets
    const estimatedHeight = Math.min(tableData.length * 36 + 50, 800);
    const tableHeight = showCollapse && !isExpanded ? Math.min(maxHeight, estimatedHeight) : estimatedHeight;

    return (
        <div className={cn("rounded-md border bg-card", className)}>
             <TableVirtuoso
                style={{ height: tableHeight }}
                data={tableData}
                initialItemCount={50}
                components={VirtuosoTableComponents}
                fixedHeaderContent={() => (
                   <TableRow>
                        {columns.map(col => (
                            <TableHead key={col} className="whitespace-nowrap font-medium text-xs px-2 py-1 h-8 bg-muted/50">
                                {col}
                            </TableHead>
                        ))}
                   </TableRow>
                )}
                itemContent={(index, row: Record<string, unknown>) => {
                    if (!row) return <></>;
                    return (
                    <>
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
                    </>
                )}}
             />

             <div className="bg-muted/30 px-2 py-1 text-[10px] text-muted-foreground border-t flex justify-between items-center">
                <span>Showing {tableData.length} rows</span>
                 {showCollapse && (
                     <Button
                        variant="ghost"
                        size="sm"
                        className="h-4 p-0 text-[10px] text-primary"
                        onClick={() => setIsExpanded(!isExpanded)}
                    >
                        {isExpanded ? "Collapse" : "Expand Viewport"}
                    </Button>
                 )}
            </div>
        </div>
    );
  };

  if (data === undefined || data === null) {
      return <span className="text-muted-foreground italic text-xs">null</span>;
  }

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
                         <TableIcon className="size-3" /> Table
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

        {viewMode === "smart" && hasSmartView ? renderSmart() : renderRaw()}
    </div>
  );
}
