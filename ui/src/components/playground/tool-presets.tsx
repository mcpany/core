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
import { Bookmark, Trash2, Save, Plus, Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient, ToolPreset } from "@/lib/client";

interface ToolPresetsProps {
  toolName: string;
  currentData: Record<string, unknown>;
  onSelect: (data: Record<string, unknown>) => void;
  serviceId?: string; // Optional context
}

export function ToolPresets({ toolName, currentData, onSelect, serviceId }: ToolPresetsProps) {
  const [presets, setPresets] = useState<ToolPreset[]>([]);
  const [loading, setLoading] = useState(false);
  const [newPresetName, setNewPresetName] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [isSaveMode, setIsSaveMode] = useState(false);
  const { toast } = useToast();

  const fetchPresets = async () => {
      setLoading(true);
      try {
          const allPresets = await apiClient.listToolPresets();
          // Filter by tool name
          setPresets(allPresets.filter(p => p.toolName === toolName));
      } catch (e) {
          console.error("Failed to load presets", e);
          toast({ variant: "destructive", title: "Failed to load presets" });
      } finally {
          setLoading(false);
      }
  };

  useEffect(() => {
    if (isOpen) {
        fetchPresets();
    }
  }, [isOpen, toolName]);

  const savePreset = async () => {
    if (!newPresetName.trim()) return;

    try {
        await apiClient.createToolPreset({
            id: crypto.randomUUID(),
            name: newPresetName.trim(),
            toolName: toolName,
            arguments: JSON.stringify(currentData),
            serviceId: serviceId
        });
        toast({ title: "Preset saved" });
        setNewPresetName("");
        setIsSaveMode(false);
        fetchPresets();
    } catch (e) {
        console.error("Failed to save preset", e);
        toast({ variant: "destructive", title: "Failed to save preset" });
    }
  };

  const deletePreset = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm("Are you sure you want to delete this preset?")) return;

    try {
        await apiClient.deleteToolPreset(id);
        toast({ title: "Preset deleted" });
        fetchPresets();
    } catch (e) {
        console.error("Failed to delete preset", e);
        toast({ variant: "destructive", title: "Failed to delete preset" });
    }
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
          {loading ? (
              <div className="flex items-center justify-center h-full text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin mr-2" /> Loading...
              </div>
          ) : presets.length === 0 ? (
            <div className="p-4 text-center text-xs text-muted-foreground">
              No saved presets for this tool.
            </div>
          ) : (
            <div className="flex flex-col p-1">
              {presets.map((preset) => (
                <div
                  key={preset.id}
                  className="flex items-center justify-between p-2 hover:bg-muted rounded-md cursor-pointer group transition-colors"
                  onClick={() => {
                      try {
                          const args = JSON.parse(preset.arguments);
                          onSelect(args);
                          setIsOpen(false);
                          toast({ title: `Loaded preset: ${preset.name}` });
                      } catch (e) {
                          toast({ variant: "destructive", title: "Failed to parse preset arguments" });
                      }
                  }}
                >
                  <span className="text-sm truncate max-w-[180px]">{preset.name}</span>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground hover:text-destructive"
                    onClick={(e) => deletePreset(preset.id, e)}
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
