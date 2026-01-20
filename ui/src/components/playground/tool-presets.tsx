/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Bookmark, Trash2, Save, Plus } from "lucide-react";
import { toast } from "sonner";

interface Preset {
  name: string;
  data: Record<string, unknown>;
}

interface ToolPresetsProps {
  toolName: string;
  currentData: Record<string, unknown>;
  onSelect: (data: Record<string, unknown>) => void;
}

export function ToolPresets({ toolName, currentData, onSelect }: ToolPresetsProps) {
  const [presets, setPresets] = useState<Preset[]>([]);
  const [newPresetName, setNewPresetName] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [isSaveMode, setIsSaveMode] = useState(false);

  const storageKey = `mcpany-presets-${toolName}`;

  // Load presets on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(storageKey);
      if (stored) {
        setPresets(JSON.parse(stored));
      }
    } catch (e) {
      console.error("Failed to load presets", e);
    }
  }, [storageKey]);

  const savePreset = () => {
    if (!newPresetName.trim()) return;

    const newPreset: Preset = {
      name: newPresetName.trim(),
      data: currentData,
    };

    // Check if exists
    const existingIndex = presets.findIndex((p) => p.name === newPreset.name);
    let updatedPresets;

    if (existingIndex >= 0) {
      if (!confirm(`Overwrite existing preset "${newPreset.name}"?`)) {
          return;
      }
      updatedPresets = [...presets];
      updatedPresets[existingIndex] = newPreset;
    } else {
      updatedPresets = [...presets, newPreset];
    }

    setPresets(updatedPresets);
    localStorage.setItem(storageKey, JSON.stringify(updatedPresets));
    setNewPresetName("");
    setIsSaveMode(false);
    toast.success("Preset saved");
  };

  const deletePreset = (name: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const updatedPresets = presets.filter((p) => p.name !== name);
    setPresets(updatedPresets);
    localStorage.setItem(storageKey, JSON.stringify(updatedPresets));
    toast.success("Preset deleted");
  };

  return (
    <Popover open={isOpen} onOpenChange={(open) => {
        setIsOpen(open);
        if (!open) setIsSaveMode(false);
    }}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="sm" className="h-8 w-8 p-0" title="Manage Presets">
          <Bookmark className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="end">
        <div className="flex items-center justify-between p-3 border-b bg-muted/30">
          <h4 className="font-medium text-sm">Presets</h4>
          {!isSaveMode && (
              <Button variant="ghost" size="icon" className="h-6 w-6" onClick={() => setIsSaveMode(true)} title="Create New Preset">
                  <Plus className="h-4 w-4" />
              </Button>
          )}
        </div>

        {isSaveMode && (
             <div className="p-3 border-b bg-muted/50">
                <div className="flex gap-2">
                    <Input
                        value={newPresetName}
                        onChange={(e) => setNewPresetName(e.target.value)}
                        placeholder="Preset Name"
                        className="h-8 text-xs"
                        autoFocus
                        onKeyDown={(e) => {
                            if (e.key === 'Enter') savePreset();
                            if (e.key === 'Escape') setIsSaveMode(false);
                        }}
                    />
                    <Button size="sm" className="h-8 w-8 p-0" onClick={savePreset} disabled={!newPresetName.trim()}>
                        <Save className="h-4 w-4" />
                    </Button>
                </div>
             </div>
        )}

        <ScrollArea className="h-[200px]">
          {presets.length === 0 ? (
            <div className="p-4 text-center text-xs text-muted-foreground">
              No saved presets for this tool.
            </div>
          ) : (
            <div className="flex flex-col p-1">
              {presets.map((preset) => (
                <div
                  key={preset.name}
                  className="flex items-center justify-between p-2 hover:bg-muted rounded-md cursor-pointer group transition-colors"
                  onClick={() => {
                      onSelect(preset.data);
                      setIsOpen(false);
                      toast.success(`Loaded preset: ${preset.name}`);
                  }}
                >
                  <span className="text-sm truncate max-w-[180px]">{preset.name}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground hover:text-destructive"
                    onClick={(e) => deletePreset(preset.name, e)}
                    title="Delete Preset"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </PopoverContent>
    </Popover>
  );
}
