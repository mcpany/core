/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useMemo, useRef, useCallback, useEffect } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { ChevronDown, ChevronUp, ChevronsUpDown, Eye, Expand, Copy, Check } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { JsonView } from "@/components/ui/json-view";
import { Input } from "@/components/ui/input";
import { Search, Database } from "lucide-react";
import { cn } from "@/lib/utils";

interface SmartTableProps {
  data: any[];
}

type SortDirection = 'asc' | 'desc' | null;

/**
 * Intent: Document SmartTable
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
 * SmartTable renders a table with sorting and filtering.
 *
 * @param props - The component props.
 * @param props.data - The data to display.
 * @returns The rendered component.
 */
export function SmartTable({ data }: SmartTableProps) {
  const [sortConfig, setSortConfig] = useState<{ key: string; direction: SortDirection }>({ key: '', direction: null });
  const [globalFilter, setGlobalFilter] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10; // Better for tool results default

  // Resizing state
  const [columnWidths, setColumnWidths] = useState<Record<string, number>>({});
  const resizingColRef = useRef<string | null>(null);
  const startXRef = useRef<number>(0);
  const startWidthRef = useRef<number>(0);

  // Extract columns
  const columns = useMemo(() => {
    const keys = new Set<string>();
    data.slice(0, 100).forEach((item) => {
      if (typeof item === 'object' && item !== null) {
        Object.keys(item).forEach(k => keys.add(k));
      }
    });
    return Array.from(keys);
  }, [data]);

  // Filtering logic
  const filteredData = useMemo(() => {
    if (!globalFilter) return data;
    const lowercasedFilter = globalFilter.toLowerCase();

    return data.filter(item => {
      return Object.values(item).some(value => {
        if (value == null) return false;
        return String(value).toLowerCase().includes(lowercasedFilter);
      });
    });
  }, [data, globalFilter]);

  // Sorting logic
  const sortedData = useMemo(() => {
    if (!sortConfig.key || !sortConfig.direction) return filteredData;

    return [...filteredData].sort((a, b) => {
      const aVal = a[sortConfig.key];
      const bVal = b[sortConfig.key];

      if (aVal === bVal) return 0;
      if (aVal == null) return 1;
      if (bVal == null) return -1;
      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortConfig.direction === 'asc' ? aVal - bVal : bVal - aVal;
      }
      const aStr = String(aVal).toLowerCase();
      const bStr = String(bVal).toLowerCase();
      if (aStr < bStr) return sortConfig.direction === 'asc' ? -1 : 1;
      if (aStr > bStr) return sortConfig.direction === 'asc' ? 1 : -1;
      return 0;
    });
  }, [filteredData, sortConfig]);

  // Pagination logic
  const totalPages = Math.ceil(sortedData.length / itemsPerPage);
  const paginatedData = useMemo(() => {
    const start = (currentPage - 1) * itemsPerPage;
    return sortedData.slice(start, start + itemsPerPage);
  }, [sortedData, currentPage, itemsPerPage]);

  const handleSort = (key: string) => {
    // Disable sorting if we are currently resizing
    if (resizingColRef.current) return;
    setSortConfig(prev => {
      if (prev.key === key) {
        if (prev.direction === 'asc') return { key, direction: 'desc' };
        if (prev.direction === 'desc') return { key: '', direction: null };
      }
      return { key, direction: 'asc' };
    });
  };

  // Drag handling
  const handleMouseDown = (e: React.MouseEvent, colKey: string) => {
    e.preventDefault();
    e.stopPropagation(); // Prevent sorting
    resizingColRef.current = colKey;
    startXRef.current = e.clientX;
    // Default width is typically 150-200 if not set, let's use current width or default
    startWidthRef.current = columnWidths[colKey] || 150;

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
  };

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!resizingColRef.current) return;

    const diff = e.clientX - startXRef.current;
    let newWidth = startWidthRef.current + diff;

    // Set a minimum width
    if (newWidth < 60) newWidth = 60;

    setColumnWidths(prev => ({
        ...prev,
        [resizingColRef.current as string]: newWidth
    }));
  }, []);

  const handleMouseUp = useCallback(() => {
    resizingColRef.current = null;
    document.removeEventListener("mousemove", handleMouseMove);
    document.removeEventListener("mouseup", handleMouseUp);
  }, [handleMouseMove]);

  useEffect(() => {
    // Cleanup event listeners on unmount
    return () => {
        document.removeEventListener("mousemove", handleMouseMove);
        document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [handleMouseMove, handleMouseUp]);


  const [copiedCell, setCopiedCell] = useState<string | null>(null);
  const copyToClipboard = (text: string, id: string) => {
      navigator.clipboard.writeText(text);
      setCopiedCell(id);
      setTimeout(() => setCopiedCell(null), 2000);
  };

  const renderCell = (value: any, colKey: string, rowIndex: number) => {
    const cellId = `${rowIndex}-${colKey}`;

    if (value === null || value === undefined) {
      return (
        <span className="inline-flex items-center text-muted-foreground/40 italic text-xs px-1.5 py-0.5 rounded-md bg-muted/20 border border-muted/30">
          null
        </span>
      );
    }
    if (typeof value === 'boolean') {
      return (
        <Badge variant={value ? "default" : "secondary"} className={cn("text-[10px] font-medium uppercase tracking-wider", value ? "bg-green-500/10 text-green-700 dark:text-green-400 hover:bg-green-500/20" : "")}>
          {String(value)}
        </Badge>
      );
    }
    if (typeof value === 'number') {
        return <span className="font-mono text-sm tabular-nums text-blue-600 dark:text-blue-400">{value}</span>;
    }
    if (typeof value === 'object') {
        const isArray = Array.isArray(value);
        const label = isArray ? `[ ] Array(${value.length})` : `{ } Object`;
        return (
            <Dialog>
                <DialogTrigger asChild>
                    <Button variant="outline" size="sm" className="h-6 px-2.5 text-[11px] font-mono rounded-full bg-muted/20 hover:bg-muted/50 text-muted-foreground hover:text-foreground border-dashed transition-all group">
                        <Eye className="mr-1.5 h-3 w-3 opacity-50 group-hover:opacity-100" />
                        {label}
                    </Button>
                </DialogTrigger>
                <DialogContent className="max-w-3xl max-h-[80vh] flex flex-col p-0 overflow-hidden bg-background/95 backdrop-blur-xl border-muted/50 shadow-2xl">
                    <DialogHeader className="px-4 py-3 border-b bg-muted/10">
                        <DialogTitle className="font-mono text-sm">{colKey}</DialogTitle>
                    </DialogHeader>
                    <ScrollArea className="flex-1 p-4 bg-black/5">
                        <JsonView data={value} />
                    </ScrollArea>
                </DialogContent>
            </Dialog>
        );
    }

    const strValue = String(value);
    if (strValue.startsWith('http://') || strValue.startsWith('https://')) {
        return (
            <a href={strValue} target="_blank" rel="noopener noreferrer" className="text-primary hover:underline text-sm truncate max-w-[250px] inline-block align-bottom">
                {strValue}
            </a>
        );
    }

    if (strValue.length > 50) {
        return (
            <div className="flex items-center gap-2 group min-w-0 relative">
                <span className="text-sm truncate flex-1 min-w-0" title={strValue}>
                    {strValue}
                </span>
                <Popover>
                    <PopoverTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity shrink-0 bg-background/80 backdrop-blur-sm border shadow-sm">
                            <Expand className="h-3 w-3 text-muted-foreground" />
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent align="end" className="w-[400px] p-0 overflow-hidden shadow-xl border-muted/60 bg-background/95 backdrop-blur-xl">
                        <div className="flex items-center justify-between px-3 py-2 border-b bg-muted/30 backdrop-blur-md">
                            <span className="text-xs font-medium text-muted-foreground font-mono">{colKey}</span>
                            <Button size="icon" variant="ghost" className="h-6 w-6 hover:bg-background/80" onClick={() => copyToClipboard(strValue, cellId)}>
                                {copiedCell === cellId ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3 text-muted-foreground" />}
                            </Button>
                        </div>
                        <ScrollArea className="max-h-[300px] w-full">
                            <div className="p-4 text-sm whitespace-pre-wrap break-words font-mono leading-relaxed bg-black/5 dark:bg-white/5">
                                {strValue}
                            </div>
                        </ScrollArea>
                    </PopoverContent>
                </Popover>
            </div>
        );
    }

    return <span className="text-sm">{strValue}</span>;
  };

  return (
    <div className="flex flex-col space-y-4 h-full">
        <div className="flex items-center">
            <div className="relative max-w-sm w-full">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search all columns..."
                    value={globalFilter}
                    onChange={(e) => {
                        setGlobalFilter(e.target.value);
                        setCurrentPage(1); // Reset to first page on search
                    }}
                    className="pl-8 h-9"
                />
            </div>
        </div>
        <div className="rounded-lg border border-border/50 bg-card overflow-hidden shadow-sm flex-1 flex flex-col">
            <ScrollArea className="flex-1 w-full relative">
                <Table>
                    <TableHeader className="bg-muted/30 sticky top-0 backdrop-blur-md z-10 shadow-sm">
                        <TableRow className="hover:bg-transparent">
                            {columns.map(col => (
                                <TableHead
                                    key={col}
                                    className="whitespace-nowrap h-10 cursor-pointer select-none transition-colors hover:bg-muted/50 px-4 group relative"
                                    onClick={() => handleSort(col)}
                                    style={{ width: columnWidths[col] ? `${columnWidths[col]}px` : 'auto', minWidth: columnWidths[col] ? `${columnWidths[col]}px` : '100px' }}
                                >
                                    <div className="flex items-center justify-between overflow-hidden">
                                        <span className="font-medium truncate mr-2" title={col}>{col}</span>
                                        <span className="text-muted-foreground/30 group-hover:text-muted-foreground flex items-center flex-shrink-0">
                                            {sortConfig.key === col ? (
                                                sortConfig.direction === 'asc' ? <ChevronUp className="h-3.5 w-3.5 text-primary" /> : <ChevronDown className="h-3.5 w-3.5 text-primary" />
                                            ) : (
                                                <ChevronsUpDown className="h-3.5 w-3.5 opacity-0 group-hover:opacity-100 transition-opacity" />
                                            )}
                                        </span>
                                    </div>
                                    <div
                                        className="absolute right-0 top-0 h-full w-2 cursor-col-resize hover:bg-primary/50 opacity-0 group-hover:opacity-100 transition-colors z-20"
                                        onMouseDown={(e) => handleMouseDown(e, col)}
                                    />
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {paginatedData.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={columns.length} className="h-48 text-center">
                                    <div className="flex flex-col items-center justify-center text-muted-foreground space-y-3">
                                        <div className="p-3 bg-muted/20 rounded-full border border-dashed border-muted">
                                            <Database className="h-6 w-6 opacity-50" />
                                        </div>
                                        <p className="text-sm">No data available.</p>
                                    </div>
                                </TableCell>
                            </TableRow>
                        ) : (
                            paginatedData.map((row, i) => (
                                <TableRow key={i} className="hover:bg-muted/30 transition-colors group/row">
                                    {columns.map(col => (
                                        <TableCell
                                            key={col}
                                            className="py-2.5 px-4"
                                            style={{ width: columnWidths[col] ? `${columnWidths[col]}px` : 'auto', maxWidth: columnWidths[col] ? `${columnWidths[col]}px` : undefined, overflow: columnWidths[col] ? 'hidden' : undefined }}
                                        >
                                            {renderCell(row[col], col, i)}
                                        </TableCell>
                                    ))}
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
                <ScrollBar orientation="horizontal" />
            </ScrollArea>
        </div>

        {totalPages > 1 && (
            <div className="flex items-center justify-between px-2 text-sm text-muted-foreground mt-2">
                <div>
                    Showing <span className="font-medium text-foreground">{(currentPage - 1) * itemsPerPage + 1}</span> to <span className="font-medium text-foreground">{Math.min(currentPage * itemsPerPage, sortedData.length)}</span> of <span className="font-medium text-foreground">{sortedData.length}</span> entries
                </div>
                <div className="flex items-center space-x-2">
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                        disabled={currentPage === 1}
                        className="h-8"
                    >
                        Previous
                    </Button>
                    <div className="flex items-center gap-1">
                        {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                            let pageNum = i + 1;
                            if (totalPages > 5 && currentPage > 3) {
                                pageNum = currentPage - 2 + i;
                                if (pageNum > totalPages) {
                                    pageNum = totalPages - (4 - i);
                                }
                            }
                            return (
                                <Button
                                    key={pageNum}
                                    variant={currentPage === pageNum ? "default" : "ghost"}
                                    size="icon"
                                    className="h-8 w-8 text-xs font-medium"
                                    onClick={() => setCurrentPage(pageNum)}
                                >
                                    {pageNum}
                                </Button>
                            );
                        })}
                    </div>
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                        disabled={currentPage === totalPages}
                        className="h-8"
                    >
                        Next
                    </Button>
                </div>
            </div>
        )}
    </div>
  );
}
