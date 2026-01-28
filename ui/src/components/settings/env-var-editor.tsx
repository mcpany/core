/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Trash2 } from "lucide-react";

export interface EnvVar {
  key: string;
  value: string;
}

interface EnvVarEditorProps {
  value: Record<string, string>;
  onChange: (value: Record<string, string>) => void;
  readOnly?: boolean;
}

export function EnvVarEditor({ value, onChange, readOnly }: EnvVarEditorProps) {
  const [entries, setEntries] = useState<EnvVar[]>(
    Object.entries(value || {}).map(([key, val]) => ({ key, value: val }))
  );

  const updateEntries = (newEntries: EnvVar[]) => {
    setEntries(newEntries);
    const newRecord: Record<string, string> = {};
    newEntries.forEach((e) => {
      if (e.key) newRecord[e.key] = e.value;
    });
    onChange(newRecord);
  };

  const addEntry = () => {
    updateEntries([...entries, { key: "", value: "" }]);
  };

  const removeEntry = (index: number) => {
    const newEntries = [...entries];
    newEntries.splice(index, 1);
    updateEntries(newEntries);
  };

  const updateEntry = (index: number, field: "key" | "value", val: string) => {
    const newEntries = [...entries];
    newEntries[index][field] = val;
    updateEntries(newEntries);
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-sm font-medium text-muted-foreground">
        <span>Key</span>
        <span>Value</span>
      </div>
      <div className="space-y-2">
        {entries.map((entry, index) => (
          <div key={index} className="flex items-center gap-2">
            <Input
              placeholder="Key"
              value={entry.key}
              onChange={(e) => updateEntry(index, "key", e.target.value)}
              disabled={readOnly}
              className="flex-1 font-mono text-xs"
            />
            <Input
              placeholder="Value"
              value={entry.value}
              onChange={(e) => updateEntry(index, "value", e.target.value)}
              disabled={readOnly}
              className="flex-1 font-mono text-xs"
            />
            {!readOnly && (
              <Button
                variant="ghost"
                size="icon"
                onClick={() => removeEntry(index)}
                className="h-8 w-8 text-muted-foreground hover:text-destructive"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            )}
          </div>
        ))}
      </div>
      {!readOnly && (
        <Button
          variant="outline"
          size="sm"
          onClick={addEntry}
          className="w-full border-dashed"
        >
          <Plus className="mr-2 h-4 w-4" /> Add Variable
        </Button>
      )}
    </div>
  );
}
