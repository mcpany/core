/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useRef } from "react";
import Editor, { useMonaco, OnMount } from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { defineDraculaTheme } from "@/lib/monaco-theme";

interface ConfigDiffViewerProps {
  diff: string;
  height?: string;
}

/**
 * ConfigDiffViewer component.
 * Displays a diff string using Monaco Editor with diff syntax highlighting.
 *
 * @param props - The component props.
 * @param props.diff - The diff string to display.
 * @param props.height - Optional height for the editor container.
 * @returns The rendered component.
 */
export function ConfigDiffViewer({ diff, height = "300px" }: ConfigDiffViewerProps) {
  const { theme, systemTheme } = useTheme();
  const monaco = useMonaco();
  const editorRef = useRef<Parameters<OnMount>[0] | null>(null);

  // Calculate actual theme
  const currentTheme = theme === "system" ? systemTheme : theme;
  const isDark = currentTheme === "dark";
  const editorTheme = isDark ? "dracula" : "light";

  const handleEditorDidMount: OnMount = (editor, monaco) => {
    editorRef.current = editor;
    if (isDark) {
      defineDraculaTheme(monaco);
      monaco.editor.setTheme("dracula");
    }
  };

  return (
    <div className="w-full overflow-hidden rounded-md border border-input bg-muted/20 shadow-sm" style={{ height }}>
      <Editor
        height="100%"
        defaultLanguage="diff"
        value={diff}
        theme={editorTheme}
        onMount={handleEditorDidMount}
        options={{
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          fontSize: 12,
          fontFamily: "var(--font-mono), monospace",
          lineNumbers: "on",
          roundedSelection: true,
          readOnly: true,
          automaticLayout: true,
          padding: { top: 16, bottom: 16 },
          renderLineHighlight: "all",
          folding: false,
          domReadOnly: true,
        }}
        loading={<div className="flex items-center justify-center h-full text-muted-foreground text-xs">Loading Diff View...</div>}
      />
    </div>
  );
}
