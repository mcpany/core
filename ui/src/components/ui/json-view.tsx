/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Code, Table as TableIcon, Copy, Check, ChevronDown, ChevronUp, ListTree } from "lucide-react";
import dynamic from "next/dynamic";
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { cn } from "@/lib/utils";
import { useToast } from "@/hooks/use-toast";
import { JsonTree } from "./json-tree";

// ⚡ BOLT: Lazy load SyntaxHighlighter to reduce initial bundle size.
// Randomized Selection from Top 5 High-Impact Targets (Assets/Bundle)
const SyntaxHighlighter = dynamic(() => import('react-syntax-highlighter').then(mod => mod.Prism), {
  ssr: false,
  loading: () => <div className="p-4 text-xs font-mono text-muted-foreground">Loading source...</div>,
});

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

// Helper to safely parse JSON if string
const tryParse = (data: unknown) => {
    if (typeof data === 'string') {
        try {
            return JSON.parse(data);
        } catch {
            return null;
        }
    }
    return data;
};

// Helper to determine table data
const getTableData = (data: unknown, smartTable: boolean) => {
    if (!smartTable) return null;
    const content = tryParse(data);

    if (Array.isArray(content) && content.length > 0) {
        const isListOfObjects = content.every(item => typeof item === 'object' && item !== null && !Array.isArray(item));
        if (isListOfObjects) {
            return content;
        }
    }
    return null;
};

/**
 * JsonView component.
 * Renders data with interactive tree view, optional smart table view, and raw syntax highlighting.
 *
 * @param props - The component props.
 * @param props.data - The data to display.
 * @param props.className - The className.
 * @param props.smartTable - Whether to attempt smart table rendering.
 * @param props.maxHeight - Max height before collapsing (only applies to Raw/Table views, Tree handles its own).
 * @returns The rendered component.
 */
export function JsonView({ data, className, smartTable = false, maxHeight = 400 }: JsonViewProps) {
  // Calculate initial state lazily
  const [viewMode, setViewMode] = useState<"smart" | "tree" | "raw">(() => {
      const tableData = getTableData(data, smartTable);
      if (tableData) return "smart";

      const parsed = tryParse(data);
      const isObj = typeof parsed === 'object' && parsed !== null;
      if (isObj) return "tree";

      return "raw";
  });

  const [copied, setCopied] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const { toast } = useToast();

  const tableData = useMemo(() => getTableData(data, smartTable), [data, smartTable]);
  const parsedData = useMemo(() => tryParse(data), [data]);

  const hasSmartView = tableData !== null;
  const isObject = typeof parsedData === 'object' && parsedData !== null;

  // Set default view mode based on data type updates
  useEffect(() => {
      if (hasSmartView) {
          setViewMode("smart");
      } else if (isObject) {
          setViewMode("tree");
      } else {
          setViewMode("raw");
      }
  }, [hasSmartView, isObject]);

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

  const showCollapse = maxHeight > 0;

  const renderCollapseButton = () => (
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
  );

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

  const renderTree = () => {
      // parsedData is already available via useMemo
      return (
          <div className={cn("rounded-md border bg-[#1e1e1e]", className)}>
              <div
                className="p-4 overflow-auto transition-all"
                style={{ maxHeight: showCollapse && !isExpanded ? `${maxHeight}px` : undefined }}
              >
                  <JsonTree data={parsedData} defaultExpandedLevel={1} />

                  {showCollapse && !isExpanded && (
                        <div className="absolute bottom-0 left-0 right-0 h-12 bg-gradient-to-t from-[#1e1e1e] to-transparent pointer-events-none" />
                  )}
              </div>
              {showCollapse && renderCollapseButton()}
          </div>
      );
  }

  const renderSmart = () => {
    if (!tableData) return renderRaw();

    // Determine columns from all keys in the first 10 rows
    const allKeys = new Set<string>();
    tableData.slice(0, 10).forEach((row: Record<string, unknown>) => {
        Object.keys(row).forEach(k => allKeys.add(k));
    });
    const columns = Array.from(allKeys);

    return (
        <div className={cn("rounded-md border overflow-hidden bg-card", className)}>
             <div
                className={cn("overflow-auto", showCollapse && !isExpanded ? "relative" : "")}
                style={{ maxHeight: showCollapse && !isExpanded ? `${maxHeight}px` : undefined }}
             >
                <Table>
                    <TableHeader className="bg-muted/50 sticky top-0 z-10">
                        <TableRow>
                            {columns.map(col => (
                                <TableHead key={col} className="whitespace-nowrap font-medium text-xs px-2 py-1 h-8 bg-muted/50">
                                    {col}
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {tableData.map((row: Record<string, unknown>, idx: number) => (
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

                 {showCollapse && !isExpanded && (
                    <div className="absolute bottom-0 left-0 right-0 h-12 bg-gradient-to-t from-background to-transparent pointer-events-none" />
                )}
            </div>

             <div className="bg-muted/30 px-2 py-1 text-[10px] text-muted-foreground border-t flex justify-between items-center">
                <span>Showing {tableData.length} rows</span>
                 {showCollapse && (
                     <Button
                        variant="ghost"
                        size="sm"
                        className="h-4 p-0 text-[10px] text-primary"
                        onClick={() => setIsExpanded(!isExpanded)}
                    >
                        {isExpanded ? "Collapse" : "Expand"}
                    </Button>
                 )}
            </div>
        </div>
    );
  };

  if (data === undefined || data === null) {
      return <span className="text-muted-foreground italic text-xs">null</span>;
  }

  // Show toolbar if we have options
  const showToolbar = hasSmartView || isObject;

  return (
    <div className="flex flex-col gap-0 w-full relative">
        {showToolbar && (
            <div className="flex justify-end mb-1 px-1">
                 <div className="flex items-center bg-muted/50 rounded-lg p-0.5 border backdrop-blur-sm">
                     {hasSmartView && (
                         <Button
                            variant={viewMode === "smart" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("smart")}
                         >
                             <TableIcon className="size-3" /> Table
                         </Button>
                     )}
                     {isObject && (
                        <Button
                            variant={viewMode === "tree" ? "secondary" : "ghost"}
                            size="sm"
                            className="h-6 px-2 text-[10px] gap-1"
                            onClick={() => setViewMode("tree")}
                         >
                             <ListTree className="size-3" /> Tree
                         </Button>
                     )}
                     <Button
                        variant={viewMode === "raw" ? "secondary" : "ghost"}
                        size="sm"
                        className="h-6 px-2 text-[10px] gap-1"
                        onClick={() => setViewMode("raw")}
                     >
                         <Code className="size-3" /> Raw
                     </Button>
                 </div>
            </div>
        )}

        <div className="mt-0">
            {viewMode === "smart" && hasSmartView ? renderSmart() :
             viewMode === "tree" ? renderTree() :
             renderRaw()}
        </div>
    </div>
  );
}
