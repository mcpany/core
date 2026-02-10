/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import Editor from "@monaco-editor/react";
import { Button } from "@/components/ui/button";
import { Save, Ban } from "lucide-react";

interface StackEditorProps {
  initialYaml?: string;
  onSave: (yaml: string) => Promise<void>;
  onCancel: () => void;
  loading?: boolean;
}

const DEFAULT_YAML = `name: new-stack
description: A new MCP stack
services:
  - name: example-service
    httpService:
      address: http://localhost:8080
`;

/**
 * StackEditor component for editing stack configurations in YAML.
 * @param props - Component properties.
 * @returns The rendered editor.
 */
export function StackEditor({ initialYaml, onSave, onCancel, loading }: StackEditorProps) {
  const [yaml, setYaml] = useState(initialYaml || DEFAULT_YAML);
  const [saving, setSaving] = useState(false);

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(yaml);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col h-full gap-4">
      <div className="flex items-center justify-between border-b pb-4">
        <h2 className="text-xl font-semibold">Configuration</h2>
        <div className="flex items-center gap-2">
            <Button variant="outline" onClick={onCancel} disabled={saving || loading}>
                <Ban className="mr-2 h-4 w-4" /> Cancel
            </Button>
            <Button onClick={handleSave} disabled={saving || loading}>
                <Save className="mr-2 h-4 w-4" /> {saving ? "Saving..." : "Deploy Stack"}
            </Button>
        </div>
      </div>
      <div className="flex-1 min-h-[500px] border rounded-md overflow-hidden bg-[#1e1e1e]">
         <Editor
            height="100%"
            defaultLanguage="yaml"
            theme="vs-dark"
            value={yaml}
            onChange={(value) => setYaml(value || "")}
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
