/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable @typescript-eslint/no-explicit-any */

import { useMemo } from "react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useTheme } from "next-themes";
import { defineDraculaTheme } from "@/lib/monaco-theme";
import { lazy, Suspense } from "react";
import isEqual from "lodash/isEqual";

const _DiffEditorLazy = lazy(() => import("@monaco-editor/react").then((mod) => ({ default: mod.DiffEditor })));
const DiffEditor = (props: any) => (
    <Suspense fallback={<div className="h-full w-full bg-[#1e1e1e] animate-pulse rounded-md" />}>
        <_DiffEditorLazy {...props} />
    </Suspense>
);

interface SmartDiffRendererProps {
    original: any;
    modified: any;
}

export function SmartDiffRenderer({ original, modified }: SmartDiffRendererProps) {
    const { theme } = useTheme();

    const [isOriginalTable, isModifiedTable] = useMemo(() => {
        const checkTable = (data: any) => {
            if (!Array.isArray(data) || data.length === 0) return false;
            return data.every((item: any) => typeof item === 'object' && item !== null && !Array.isArray(item));
        };
        return [checkTable(original), checkTable(modified)];
    }, [original, modified]);

    const isTableDiff = isOriginalTable && isModifiedTable;

    const tableDiffData = useMemo(() => {
        if (!isTableDiff) return null;

        const origRows = Array.isArray(original) ? original : [];
        const modRows = Array.isArray(modified) ? modified : [];

        // Simple diff strategy:
        // 1. Find unchanged rows (present in both)
        // 2. Find removed rows (in orig, not in mod)
        // 3. Find added rows (in mod, not in orig)

        const diffRows: Array<{ type: 'unchanged' | 'added' | 'removed', row: any }> = [];

        // We'll keep track of which mod rows have been matched to avoid double counting
        const matchedModIndices = new Set<number>();

        origRows.forEach((origRow) => {
            const modIndex = modRows.findIndex((modRow, i) => !matchedModIndices.has(i) && isEqual(origRow, modRow));
            if (modIndex !== -1) {
                diffRows.push({ type: 'unchanged', row: origRow });
                matchedModIndices.add(modIndex);
            } else {
                diffRows.push({ type: 'removed', row: origRow });
            }
        });

        modRows.forEach((modRow, i) => {
            if (!matchedModIndices.has(i)) {
                diffRows.push({ type: 'added', row: modRow });
            }
        });

        return diffRows;
    }, [isTableDiff, original, modified]);

    const tableColumns = useMemo(() => {
        if (!tableDiffData) return [];
        const allKeys = new Set<string>();
        tableDiffData.forEach(({ row }) => {
            if (typeof row === 'object' && row !== null) {
                Object.keys(row).forEach(k => allKeys.add(k));
            }
        });
        return Array.from(allKeys);
    }, [tableDiffData]);

    if (isTableDiff && tableDiffData) {
        return (
            <div className="flex-1 overflow-auto bg-card rounded-md border">
                <Table>
                    <TableHeader className="bg-muted/50 sticky top-0 z-10 shadow-sm">
                        <TableRow>
                            <TableHead className="w-[50px] text-center border-r font-mono text-xs text-muted-foreground p-0 m-0 leading-none">+/-</TableHead>
                            {tableColumns.map(col => (
                                <TableHead key={col} className="whitespace-nowrap font-medium text-xs px-3 py-2 h-auto">
                                    {col}
                                </TableHead>
                            ))}
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {tableDiffData.map((item, idx) => {
                            let bgClass = "";
                            let sign = " ";
                            let signColor = "text-muted-foreground";

                            if (item.type === 'added') {
                                bgClass = "bg-green-500/10 dark:bg-green-500/20";
                                sign = "+";
                                signColor = "text-green-600 dark:text-green-400 font-bold";
                            } else if (item.type === 'removed') {
                                bgClass = "bg-red-500/10 dark:bg-red-500/20";
                                sign = "-";
                                signColor = "text-red-600 dark:text-red-400 font-bold";
                            }

                            return (
                                <TableRow key={idx} className={`${bgClass} border-b border-border/50 hover:${bgClass}`}>
                                    <TableCell className={`w-[50px] text-center border-r font-mono text-sm p-0 m-0 select-none ${signColor}`}>
                                        {sign}
                                    </TableCell>
                                    {tableColumns.map(col => {
                                        const val = item.row[col];
                                        let displayVal = val;
                                        if (typeof val === 'object' && val !== null) {
                                            displayVal = JSON.stringify(val);
                                        } else if (typeof val === 'boolean') {
                                            displayVal = val ? "true" : "false";
                                        }

                                        return (
                                            <TableCell
                                                key={col}
                                                className={`px-3 py-2 text-xs truncate max-w-[300px] ${item.type === 'removed' ? 'line-through opacity-70' : ''}`}
                                                title={String(displayVal ?? "")}
                                            >
                                                {String(displayVal ?? "")}
                                            </TableCell>
                                        );
                                    })}
                                </TableRow>
                            );
                        })}
                    </TableBody>
                </Table>
            </div>
        );
    }

    return (
        <div className="flex-1 border rounded-md overflow-hidden bg-[#1e1e1e]">
            <DiffEditor
                original={JSON.stringify(original, null, 2)}
                modified={JSON.stringify(modified, null, 2)}
                language="json"
                theme={theme === "dark" ? "dracula" : "light"}
                onMount={(_editor: any, monaco: any) => {
                    if (theme === "dark") {
                        defineDraculaTheme(monaco);
                        monaco.editor.setTheme("dracula");
                    }
                }}
                options={{
                    readOnly: true,
                    minimap: { enabled: false },
                    scrollBeyondLastLine: false,
                    fontSize: 12,
                    diffCodeLens: true,
                    renderSideBySide: true,
                }}
            />
        </div>
    );
}
