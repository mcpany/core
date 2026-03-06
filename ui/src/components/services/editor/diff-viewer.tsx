/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { DiffEditor } from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { Columns, AlignLeft } from "lucide-react";

interface DiffViewerProps {
    original: string;
    modified: string;
    language?: string;
}

/**
 * DiffViewer component for comparing configurations (YAML/JSON).
 * Uses Monaco Diff Editor.
 *
 * @param props - The component props.
 * @param props.original - The original content string.
 * @param props.modified - The modified content string.
 * @param props.language - The language for syntax highlighting (default: "yaml").
 * @returns The rendered diff editor.
 */
export function DiffViewer({ original, modified, language = "yaml" }: DiffViewerProps) {
    const { theme } = useTheme();
    const [isSideBySide, setIsSideBySide] = useState(false);

    return (
        <div className="flex flex-col h-[500px] border rounded-md overflow-hidden bg-background">
            <div className="flex justify-end p-2 border-b bg-muted/20">
                <div className="flex bg-muted rounded-md p-1">
                    <Button
                        variant={!isSideBySide ? "secondary" : "ghost"}
                        size="sm"
                        onClick={() => setIsSideBySide(false)}
                        className="h-7 px-2 text-xs"
                    >
                        <AlignLeft className="mr-1 h-3 w-3" />
                        Inline
                    </Button>
                    <Button
                        variant={isSideBySide ? "secondary" : "ghost"}
                        size="sm"
                        onClick={() => setIsSideBySide(true)}
                        className="h-7 px-2 text-xs"
                    >
                        <Columns className="mr-1 h-3 w-3" />
                        Side-by-Side
                    </Button>
                </div>
            </div>
            <div className="flex-1 min-h-0">
                <DiffEditor
                    height="100%"
                    language={language}
                    original={original}
                    modified={modified}
                    theme={theme === "dark" ? "vs-dark" : "light"}
                    options={{
                        minimap: { enabled: false },
                        scrollBeyondLastLine: false,
                        fontSize: 12,
                        wordWrap: "on",
                        readOnly: true,
                        renderSideBySide: isSideBySide
                    }}
                />
            </div>
        </div>
    );
}
