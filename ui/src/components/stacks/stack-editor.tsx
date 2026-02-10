/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle, Save, X } from "lucide-react";
import yaml from "js-yaml";
import { ConfigEditor } from "./config-editor";

interface StackEditorProps {
  initialContent?: string;
  onSave: (content: string) => Promise<void>;
  onCancel: () => void;
}

/**
 * StackEditor component for editing stack configurations (YAML).
 * @param props The component props.
 * @returns The rendered component.
 */
export function StackEditor({ initialContent = "", onSave, onCancel }: StackEditorProps) {
  const [content, setContent] = useState(initialContent);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const handleSave = async () => {
    setError(null);
    setSaving(true);
    try {
        // Basic YAML validation
        try {
            yaml.load(content);
        } catch (e: any) {
            throw new Error(`Invalid YAML: ${e.message}`);
        }

        await onSave(content);
    } catch (e: any) {
        setError(e.message || "Failed to save stack");
    } finally {
        setSaving(false);
    }
  };

  return (
    <div className="flex flex-col h-full space-y-4">
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="flex-1 min-h-[500px]">
        <ConfigEditor
          value={content}
          onChange={(value) => setContent(value || "")}
          language="yaml"
        />
      </div>

      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={onCancel} disabled={saving}>
          <X className="mr-2 h-4 w-4" /> Cancel
        </Button>
        <Button onClick={handleSave} disabled={saving}>
          <Save className="mr-2 h-4 w-4" /> {saving ? "Saving..." : "Deploy Stack"}
        </Button>
      </div>
    </div>
  );
}
