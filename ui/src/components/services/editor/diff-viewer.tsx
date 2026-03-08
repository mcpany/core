/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import { DiffEditor } from "@monaco-editor/react";
import { useTheme } from "next-themes";

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

    return (
        <div className="h-full min-h-[500px] border rounded-md overflow-hidden flex flex-col">
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
                    renderSideBySide: true
                }}
            />
        </div>
    );
}
