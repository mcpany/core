/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import Editor from "@monaco-editor/react";
import { useTheme } from "next-themes";

interface SourceEditorProps {
    value: string;
    onChange: (value: string | undefined) => void;
}

/**
 * SourceEditor component for editing YAML configuration.
 * Uses Monaco Editor.
 *
 * @param props - The component props.
 * @param props.value - The current YAML string.
 * @param props.onChange - Callback when value changes.
 * @returns The rendered editor.
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
