/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect, useRef } from "react";
import Editor, { useMonaco, OnMount } from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { STACK_CONFIG_SCHEMA } from "@/lib/stack-schema";

interface ConfigEditorProps {
  value: string;
  onChange: (value: string | undefined) => void;
  language?: string;
  readOnly?: boolean;
}

/**
 * ConfigEditor.
 *
 * @param readOnly = false - The readOnly = false.
 */
export function ConfigEditor({ value, onChange, language = "yaml", readOnly = false }: ConfigEditorProps) {
  const { theme, systemTheme } = useTheme();
  const monaco = useMonaco();
  const editorRef = useRef<Parameters<OnMount>[0] | null>(null);

  // Calculate actual theme
  const currentTheme = theme === "system" ? systemTheme : theme;
  const editorTheme = currentTheme === "dark" ? "vs-dark" : "light";

  useEffect(() => {
    if (monaco && language === "yaml") {
        // Configure JSON Validation for YAML?
        // Monaco's YAML support is limited out of the box.
        // We can try to use setDiagnosticsOptions for json if we were editing JSON.
        // But for YAML, we rely on basic highlighting.
        // However, if we change language to 'json', we get schema validation.
        // Since we are strictly doing YAML, we'll skip deep schema validation *inside* Monaco
        // unless we bring in a YAML worker (complex).
        // Instead, we rely on the parent component's js-yaml validation to show errors.

        // But we CAN add custom completion providers for YAML!
        const disposable = monaco.languages.registerCompletionItemProvider("yaml", {
            provideCompletionItems: (model, position) => {
                const word = model.getWordUntilPosition(position);
                const range = {
                    startLineNumber: position.lineNumber,
                    endLineNumber: position.lineNumber,
                    startColumn: word.startColumn,
                    endColumn: word.endColumn,
                };

                // Basic suggestions based on our schema
                const suggestions = [
                    {
                        label: "services",
                        kind: monaco.languages.CompletionItemKind.Keyword,
                        insertText: "services:\n  ",
                        documentation: "Define services block",
                        range
                    },
                    {
                        label: "version",
                        kind: monaco.languages.CompletionItemKind.Keyword,
                        insertText: 'version: "1.0"',
                        documentation: "Stack configuration version",
                        range
                    },
                    {
                        label: "image",
                        kind: monaco.languages.CompletionItemKind.Property,
                        insertText: "image: ",
                        documentation: "Docker image",
                        range
                    },
                    {
                        label: "command",
                        kind: monaco.languages.CompletionItemKind.Property,
                        insertText: "command: ",
                        documentation: "Command to execute",
                        range
                    },
                    {
                        label: "environment",
                        kind: monaco.languages.CompletionItemKind.Property,
                        insertText: "environment:\n    - KEY=VALUE",
                        documentation: "Environment variables",
                        range
                    }
                ];
                return { suggestions };
            }
        });

        return () => disposable.dispose();
    }
  }, [monaco, language]);


  const handleEditorDidMount: OnMount = (editor, _monaco) => {
    editorRef.current = editor;
  };

  return (
    <div className="h-full w-full overflow-hidden rounded-md border border-input bg-transparent shadow-sm focus-within:ring-1 focus-within:ring-ring">
      <Editor
        height="100%"
        defaultLanguage={language}
        value={value}
        onChange={onChange}
        theme={editorTheme}
        onMount={handleEditorDidMount}
        options={{
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          fontSize: 13,
          fontFamily: "var(--font-mono), monospace",
          lineNumbers: "on",
          roundedSelection: true,
          readOnly,
          automaticLayout: true,
          padding: { top: 16, bottom: 16 },
          overviewRulerLanes: 0,
          renderLineHighlight: "all",
          lineDecorationsWidth: 16, // minimal width for line numbers
          folding: true,
        }}
        loading={<div className="flex items-center justify-center h-full text-muted-foreground text-xs">Loading Editor...</div>}
      />
    </div>
  );
}
