/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import { DiffEditor } from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { defineDraculaTheme } from "@/lib/monaco-theme";

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
    const { theme, systemTheme } = useTheme();
    const currentTheme = theme === "system" ? systemTheme : theme;
    const isDark = currentTheme === "dark";
    const editorTheme = isDark ? "dracula" : "light";

    return (
        <div className="h-full w-full min-h-[400px] overflow-hidden rounded-md border border-input bg-background/50 backdrop-blur-sm transition-all duration-300">
            <DiffEditor
                height="100%"
                language={language}
                original={original}
                modified={modified}
                theme={editorTheme}
                onMount={(editor, monaco) => {
                    if (isDark) {
                        defineDraculaTheme(monaco);
                        monaco.editor.setTheme("dracula");
                    }
                }}
                options={{
                    readOnly: true,
                    minimap: { enabled: false },
                    scrollBeyondLastLine: false,
                    fontSize: 13,
                    fontFamily: "var(--font-mono), monospace",
                    diffCodeLens: true,
                    renderSideBySide: true,
                    padding: { top: 16, bottom: 16 },
                    automaticLayout: true,
                }}
                loading={<div className="flex items-center justify-center h-full text-muted-foreground text-xs animate-pulse">Loading Diff...</div>}
                className="monaco-diff-editor"
            />
        </div>
    );
}
