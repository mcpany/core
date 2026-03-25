/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { ChevronDown, ChevronUp, ChevronsUpDown, Maximize2, MoreHorizontal, Copy, Check } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { JsonView } from "@/components/ui/json-view";
import { cn } from "@/lib/utils";

interface SmartTableProps {
  data: any[];
}

type SortDirection = 'asc' | 'desc' | null;

/**
 * SmartTable renders a table with sorting and filtering.
 *
 * @param props - The component props.
 * @param props.data - The data to display.
 * @returns The rendered component.
 */
export function SmartTable({ data }: SmartTableProps) {
  const [sortConfig, setSortConfig] = useState<{ key: string; direction: SortDirection }>({ key: '', direction: null });
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10; // Better for tool results default

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

  // Sorting logic
  const sortedData = useMemo(() => {
    if (!sortConfig.key || !sortConfig.direction) return data;

    return [...data].sort((a, b) => {
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
  }, [data, sortConfig]);

  // Pagination logic
  const totalPages = Math.ceil(sortedData.length / itemsPerPage);
  const paginatedData = useMemo(() => {
    const start = (currentPage - 1) * itemsPerPage;
    return sortedData.slice(start, start + itemsPerPage);
  }, [sortedData, currentPage, itemsPerPage]);

  const handleSort = (key: string) => {
    setSortConfig(prev => {
      if (prev.key === key) {
        if (prev.direction === 'asc') return { key, direction: 'desc' };
        if (prev.direction === 'desc') return { key: '', direction: null };
      }
      return { key, direction: 'asc' };
    });
  };

  const [copiedCell, setCopiedCell] = useState<string | null>(null);
  const copyToClipboard = (text: string, id: string) => {
      navigator.clipboard.writeText(text);
      setCopiedCell(id);
      setTimeout(() => setCopiedCell(null), 2000);
  };

  const renderCell = (value: any, colKey: string, rowIndex: number) => {
    const cellId = `${rowIndex}-${colKey}`;

    if (value === null || value === undefined) {
      return <span className="text-muted-foreground/50 italic text-xs">null</span>;
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
        const label = isArray ? `Array(${value.length})` : `Object {${Object.keys(value).length}}`;
        return (
            <Dialog>
                <DialogTrigger asChild>
                    <Button variant="ghost" size="sm" className="h-6 px-2 text-xs font-mono bg-muted/30 hover:bg-muted/60 text-muted-foreground border border-transparent hover:border-border">
                        <Maximize2 className="mr-1.5 h-3 w-3" />
                        {label}
                    </Button>
                </DialogTrigger>
                <DialogContent className="max-w-3xl max-h-[80vh] flex flex-col p-0 overflow-hidden bg-background/95 backdrop-blur-xl border-muted/50 shadow-2xl">
                    <DialogHeader className="px-4 py-3 border-b bg-muted/10">
                        <DialogTitle className="font-mono text-sm">{colKey}</DialogTitle>
                    </DialogHeader>
                    <ScrollArea className="flex-1 p-4">
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
            <div className="flex items-center gap-2 group">
                <span className="text-sm truncate max-w-[200px] sm:max-w-[300px]" title={strValue}>
                    {strValue}
                </span>
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
                            <MoreHorizontal className="h-3 w-3" />
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-[300px] p-0">
                        <div className="flex items-center justify-between p-2 border-b bg-muted/20">
                            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">{colKey}</span>
                            <Button size="icon" variant="ghost" className="h-6 w-6" onClick={() => copyToClipboard(strValue, cellId)}>
                                {copiedCell === cellId ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
                            </Button>
                        </div>
                        <div className="p-3 text-sm max-h-[300px] overflow-y-auto whitespace-pre-wrap break-words font-mono">
                            {strValue}
                        </div>
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
        );
    }

    return <span className="text-sm">{strValue}</span>;
  };

  return (
    <div className="flex flex-col space-y-4 h-full">
        <div className="rounded-lg border border-border/50 bg-card overflow-hidden shadow-sm flex-1 flex flex-col">
            <ScrollArea className="flex-1 w-full relative">
                <Table>
                    <TableHeader className="bg-muted/30 sticky top-0 backdrop-blur-md z-10 shadow-sm">
                        <TableRow className="hover:bg-transparent">
                            {columns.map(col => (
                                <TableHead
                                    key={col}
                                    className="whitespace-nowrap h-10 cursor-pointer select-none transition-colors hover:bg-muted/50 px-4 group"
                                    onClick={() => handleSort(col)}
                                >
                                    <div className="flex items-center justify-between">
                                        <span className="font-medium">{col}</span>
                                        <span className="text-muted-foreground/30 group-hover:text-muted-foreground flex items-center ml-2">
                                            {sortConfig.key === col ? (
                                                sortConfig.direction === 'asc' ? <ChevronUp className="h-3.5 w-3.5 text-primary" /> : <ChevronDown className="h-3.5 w-3.5 text-primary" />
                                            ) : (
                                                <ChevronsUpDown className="h-3.5 w-3.5 opacity-0 group-hover:opacity-100 transition-opacity" />
                                            )}
                                        </span>
                                    </div>
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {paginatedData.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">
                                    No data available.
                                </TableCell>
                            </TableRow>
                        ) : (
                            paginatedData.map((row, i) => (
                                <TableRow key={i} className="hover:bg-muted/30 transition-colors group/row">
                                    {columns.map(col => (
                                        <TableCell key={col} className="py-2.5 px-4">
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
