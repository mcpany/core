/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import dynamic from "next/dynamic";
import { useTheme } from "next-themes";
import yaml from "js-yaml";
import { UpstreamServiceConfig } from "@/lib/types";
import { defineDraculaTheme } from "@/lib/monaco-theme";

// ⚡ Bolt Optimization: Lazy load Monaco DiffEditor to avoid bundling it in the main chunk.
// Randomized Selection from Top 5 High-Impact Targets
const DiffEditor = dynamic(
  () => import("@monaco-editor/react").then((mod) => mod.DiffEditor),
  {
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center h-full text-muted-foreground text-xs">
        Loading Editor...
      </div>
    ),
  }
);

interface ServiceConfigDiffProps {
  original: UpstreamServiceConfig;
  modified: UpstreamServiceConfig;
}

/**
 * ServiceConfigDiff component.
 * @param props - The component props.
 * @param props.original - The original property.
 * @param props.modified - The modified property.
 * @returns The rendered component.
 */
export function ServiceConfigDiff({ original, modified }: ServiceConfigDiffProps) {
  const { theme, systemTheme } = useTheme();

  // Calculate actual theme
  // Calculate actual theme
  const currentTheme = theme === "system" ? systemTheme : theme;
  const isDark = currentTheme === "dark";
  const editorTheme = isDark ? "dracula" : "light";

  // Dump to YAML
  // We use simple sorting to ensure keys are in consistent order for better diffs
  const originalYaml = yaml.dump(original, { sortKeys: true, indent: 2, lineWidth: -1 });
  const modifiedYaml = yaml.dump(modified, { sortKeys: true, indent: 2, lineWidth: -1 });

  return (
    <div className="h-[400px] w-full overflow-hidden rounded-md border border-input bg-background">
      <DiffEditor
        height="100%"
        original={originalYaml}
        modified={modifiedYaml}
        language="yaml"
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
          renderSideBySide: true, // Side by side diff
          padding: { top: 16, bottom: 16 },
          automaticLayout: true,
          diffCodeLens: true,
        }}
        loading={<div className="flex items-center justify-center h-full text-muted-foreground text-xs">Loading Diff...</div>}
      />
    </div>
  );
}
