"use client";

import { useState, useEffect } from "react";
import { apiClient, ToolPreset } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Play, Trash2, Search, Loader2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";

interface PresetListProps {
    onSelect: (toolName: string, args: Record<string, unknown>) => void;
}

export function PresetList({ onSelect }: PresetListProps) {
    const [presets, setPresets] = useState<ToolPreset[]>([]);
    const [searchQuery, setSearchQuery] = useState("");
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const fetchPresets = async () => {
        setLoading(true);
        try {
            const data = await apiClient.listToolPresets();
            setPresets(data);
        } catch (e) {
            console.error(e);
            toast({ variant: "destructive", title: "Failed to load presets" });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchPresets();
    }, []);

    const handleDelete = async (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        if (!confirm("Are you sure?")) return;
        try {
            await apiClient.deleteToolPreset(id);
            toast({ title: "Preset deleted" });
            setPresets(prev => prev.filter(p => p.id !== id));
        } catch (e) {
            toast({ variant: "destructive", title: "Failed to delete preset" });
        }
    };

    const filtered = presets.filter(p =>
        p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        p.toolName.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="flex flex-col h-full bg-muted/10 border-r">
            <div className="p-4 border-b space-y-3">
                <div className="relative">
                    <Search className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                    <Input
                        placeholder="Search presets..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-8 h-9 text-xs"
                    />
                </div>
            </div>
            <ScrollArea className="flex-1">
                {loading ? (
                    <div className="flex justify-center p-4">
                        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                    </div>
                ) : filtered.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground text-xs">
                        No presets found.
                    </div>
                ) : (
                    <div className="p-3 space-y-2">
                        {filtered.map(preset => (
                            <div
                                key={preset.id}
                                className="group flex flex-col gap-2 p-3 rounded-lg border bg-card hover:bg-accent/50 transition-all cursor-pointer shadow-sm"
                                onClick={() => {
                                    try {
                                        const args = JSON.parse(preset.arguments);
                                        onSelect(preset.toolName, args);
                                    } catch {
                                        toast({ variant: "destructive", title: "Invalid preset arguments" });
                                    }
                                }}
                            >
                                <div className="flex items-start justify-between">
                                    <span className="font-semibold text-sm">{preset.name}</span>
                                    <Badge variant="outline" className="text-[10px] h-4 px-1">{preset.toolName}</Badge>
                                </div>
                                <div className="flex items-center justify-between pt-1 opacity-60 group-hover:opacity-100 transition-opacity">
                                    <span className="text-[10px] text-muted-foreground">
                                        Click to load
                                    </span>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 text-destructive hover:text-destructive hover:bg-destructive/10"
                                        onClick={(e) => handleDelete(preset.id, e)}
                                    >
                                        <Trash2 className="h-3 w-3" />
                                    </Button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </ScrollArea>
        </div>
    );
}
