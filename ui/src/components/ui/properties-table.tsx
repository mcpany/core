/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Table, TableBody, TableCell, TableRow } from "@/components/ui/table";
import { Copy, Check, Maximize2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { JsonView } from "@/components/ui/json-view";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface PropertiesTableProps {
    data: Record<string, any>;
    className?: string;
}

export function PropertiesTable({ data, className }: PropertiesTableProps) {
    const [copiedCell, setCopiedCell] = useState<string | null>(null);

    if (!data || typeof data !== 'object' || Array.isArray(data)) {
        return <span className="text-muted-foreground italic text-sm">Invalid data for properties view</span>;
    }

    const entries = Object.entries(data);

    if (entries.length === 0) {
        return <span className="text-muted-foreground italic text-sm">Empty object</span>;
    }

    const copyToClipboard = (text: string, id: string) => {
        navigator.clipboard.writeText(text);
        setCopiedCell(id);
        setTimeout(() => setCopiedCell(null), 2000);
    };

    const renderValue = (value: any, key: string) => {
        if (value === null || value === undefined) {
            return <span className="text-muted-foreground/50 italic text-sm">null</span>;
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
                            <DialogTitle className="font-mono text-sm">{key}</DialogTitle>
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

        return (
            <div className="flex items-center justify-between group">
                <span className="text-sm whitespace-pre-wrap break-words break-all">{strValue}</span>
                {strValue.length > 0 && (
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity shrink-0 ml-2"
                        onClick={() => copyToClipboard(strValue, key)}
                    >
                        {copiedCell === key ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3 text-muted-foreground" />}
                    </Button>
                )}
            </div>
        );
    };

    return (
        <div className={cn("rounded-md border border-border/50 bg-card overflow-hidden shadow-sm", className)}>
            <Table>
                <TableBody>
                    {entries.map(([key, value], i) => (
                        <TableRow key={key} className={cn("transition-colors hover:bg-muted/30 group/row", i % 2 === 0 ? "bg-muted/5" : "")}>
                            <TableCell className="py-3 px-4 w-1/3 align-top font-medium text-sm text-muted-foreground border-r border-border/50 bg-muted/10">
                                {key}
                            </TableCell>
                            <TableCell className="py-3 px-4 align-top">
                                {renderValue(value, key)}
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}
