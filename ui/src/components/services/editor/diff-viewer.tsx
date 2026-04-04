/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import React from "react";
import { DiffEditor } from "@monaco-editor/react";
import { useTheme } from "next-themes";

interface DiffViewerProps {
    original: string;
    modified: string;
    language?: string;
}

/**
 * DiffViewer executes the DiffViewer logic.
 *
 * Summary: Executes the DiffViewer logic.
 *
 * @param params - The parameters for the operation.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export function DiffViewer({ original, modified, language = "yaml" }: DiffViewerProps) {
    const { theme } = useTheme();

    return (
        <div className="h-[500px] border rounded-md overflow-hidden">
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
