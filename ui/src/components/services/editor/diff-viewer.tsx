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
}

/**
 * DiffViewer component for comparing YAML configurations.
 * Uses Monaco Diff Editor.
 *
 * @param props - The component props.
 * @param props.original - The original YAML string.
 * @param props.modified - The modified YAML string.
 * @returns The rendered diff editor.
 */
export function DiffViewer({ original, modified }: DiffViewerProps) {
    const { theme } = useTheme();

    return (
        <div className="h-[500px] border rounded-md overflow-hidden">
            <DiffEditor
                height="100%"
                language="yaml"
                original={original}
                modified={modified}
                theme={theme === "dark" ? "vs-dark" : "light"}
                options={{
                    minimap: { enabled: false },
                    scrollBeyondLastLine: false,
                    fontSize: 12,
                    tabSize: 2,
                    wordWrap: "on",
                    readOnly: true,
                    renderSideBySide: true
                }}
            />
        </div>
    );
}
