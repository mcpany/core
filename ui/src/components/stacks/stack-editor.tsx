/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import Editor from "@monaco-editor/react";
import { Button } from "@/components/ui/button";
import { Loader2, Save, X } from "lucide-react";
import { useTheme } from "next-themes";

interface StackEditorProps {
  initialValue: string;
  onSave: (value: string) => Promise<void>;
  onCancel: () => void;
}

/**
 * StackEditor component.
 * Provides a YAML editor for editing stack configurations.
 * @param props - The component props.
 * @param props.initialValue - The initial YAML content.
 * @param props.onSave - Callback when save is clicked.
 * @param props.onCancel - Callback when cancel is clicked.
 * @returns The rendered component.
 */
export function StackEditor({ initialValue, onSave, onCancel }: StackEditorProps) {
  const [value, setValue] = useState(initialValue);
  const [saving, setSaving] = useState(false);
  const { theme } = useTheme();

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(value);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col h-full border rounded-md overflow-hidden shadow-sm">
      <div className="flex items-center justify-between p-2 border-b bg-muted/40">
        <div className="text-sm font-medium pl-2">Configuration (YAML)</div>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" onClick={onCancel} disabled={saving}>
            <X className="mr-2 h-4 w-4" /> Cancel
          </Button>
          <Button size="sm" onClick={handleSave} disabled={saving}>
            {saving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
            Save Stack
          </Button>
        </div>
      </div>
      <div className="flex-1 min-h-0">
        <Editor
          height="100%"
          defaultLanguage="yaml"
          value={value}
          onChange={(val) => setValue(val || "")}
          theme={theme === "dark" ? "vs-dark" : "light"}
          options={{
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            fontSize: 13,
            fontFamily: "var(--font-mono)",
          }}
        />
      </div>
    </div>
  );
}
