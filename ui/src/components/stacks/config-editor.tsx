/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useRef } from "react";
import Editor, { OnMount } from "@monaco-editor/react";
import { useTheme } from "next-themes";

interface ConfigEditorProps {
    value: string;
    onChange: (value: string | undefined) => void;
    readOnly?: boolean;
}

/**
 * ConfigEditor component.
 * Wraps Monaco Editor for editing YAML configurations.
 *
 * @param props - The component props.
 * @param props.value - The current content of the editor.
 * @param props.onChange - Callback when content changes.
 * @param props.readOnly - Whether the editor is read-only.
 * @returns The rendered editor component.
 */
export function ConfigEditor({ value, onChange, readOnly = false }: ConfigEditorProps) {
    const { theme } = useTheme();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const editorRef = useRef<any>(null);

    const handleEditorDidMount: OnMount = (editor, monaco) => {
        editorRef.current = editor;

        // Configure YAML validation/schema if needed
        monaco.languages.yaml?.yamlDefaults.setDiagnosticsOptions({
            validate: true,
            enableSchemaRequest: true,
            schemas: [] // We could inject schemas here
        });
    };

    return (
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
                fontFamily: "monospace",
                readOnly: readOnly,
                wordWrap: "on",
                tabSize: 2,
            }}
            onMount={handleEditorDidMount}
        />
    );
}
