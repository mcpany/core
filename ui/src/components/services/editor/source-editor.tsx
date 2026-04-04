/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import React from "react";
import Editor from "@monaco-editor/react";
import { useTheme } from "next-themes";

interface SourceEditorProps {
    value: string;
    onChange: (value: string | undefined) => void;
}

/**
 * SourceEditor executes the SourceEditor logic.
 *
 * Summary: Executes the SourceEditor logic.
 *
 * @param params - The parameters for the operation.
 * @returns The result of the operation.
 * @throws An error if the operation fails.
 */
export function SourceEditor({ value, onChange }: SourceEditorProps) {
    const { theme } = useTheme();

    return (
        <div className="h-[500px] border rounded-md overflow-hidden">
            <Editor
                height="100%"
                defaultLanguage="yaml"
                value={value}
                onChange={onChange}
                theme={theme === "dark" ? "vs-dark" : "light"}
                options={{
                    minimap: { enabled: false },
                    scrollBeyondLastLine: false,
                    fontSize: 12,
                    tabSize: 2,
                    wordWrap: "on"
                }}
            />
        </div>
    );
}
