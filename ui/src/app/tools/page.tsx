/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Wrench, Play, Star, Search, List, LayoutList, PlayCircle, PauseCircle } from "lucide-react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { ToolInspector } from "@/components/tools/tool-inspector";
import { usePinnedTools } from "@/hooks/use-pinned-tools";
import { estimateTokens, formatTokenCount } from "@/lib/tokens";
import { Info } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

/**
 * ToolsPage component.
 * @returns The rendered component.
 */
export default function ToolsPage() {
  const [tools, setTools] = useState<ToolDefinition[]>([]);
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [selectedTool, setSelectedTool] = useState<ToolDefinition | null>(null);
  const [inspectorOpen, setInspectorOpen] = useState(false);
  const { isPinned, togglePin, isLoaded } = usePinnedTools();
  const [showPinnedOnly, setShowPinnedOnly] = useState(false);
  const [selectedService, setSelectedService] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [isCompact, setIsCompact] = useState(false);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const { toast } = useToast();

  useEffect(() => {
    const savedCompact = localStorage.getItem("tools_compact_view") === "true";
    setIsCompact(savedCompact);
  }, []);

  useEffect(() => {
    fetchTools();
    fetchServices();
  }, []);

  const fetchTools = async () => {
    try {
      const res = await apiClient.listTools();
      setTools(res.tools || []);
    } catch (e) {
      console.error("Failed to fetch tools", e);
    }
  };

  const fetchServices = async () => {
    try {
      const res = await apiClient.listServices();
      setServices(res);
    } catch (e) {
      console.error("Failed to fetch services", e);
    }
  };

  const toggleTool = async (name: string, currentStatus: boolean) => {
    // Optimistic update
    setTools(tools.map(t => t.name === name ? { ...t, disable: currentStatus } : t));

    try {
        await apiClient.setToolStatus(name, !currentStatus);
    } catch (e) {
        console.error("Failed to toggle tool", e);
        fetchTools(); // Revert
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to update tool status."
        });
    }
  };

  const toggleCompact = () => {
    const newState = !isCompact;
    setIsCompact(newState);
    localStorage.setItem("tools_compact_view", String(newState));
  };

  const openInspector = (tool: ToolDefinition) => {
      setSelectedTool(tool);
      setInspectorOpen(true);
  };

  const filteredTools = tools
    .filter((t) => !showPinnedOnly || isPinned(t.name))
    .filter((t) => selectedService === "all" || t.serviceId === selectedService)
    .filter((t) =>
      searchQuery === "" ||
      t.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      t.description.toLowerCase().includes(searchQuery.toLowerCase())
    )
    .sort((a, b) => {
      const aPinned = isPinned(a.name);
      const bPinned = isPinned(b.name);
      if (aPinned && !bPinned) return -1;
      if (!aPinned && bPinned) return 1;
      return a.name.localeCompare(b.name);
    });

  const handleSelectAll = useCallback((checked: boolean) => {
    if (checked) {
      setSelected(new Set(filteredTools.map(t => t.name)));
    } else {
      setSelected(new Set());
    }
  }, [filteredTools]);

  const handleSelectOne = useCallback((name: string, checked: boolean) => {
    setSelected(prev => {
        const newSelected = new Set(prev);
        if (checked) {
          newSelected.add(name);
        } else {
          newSelected.delete(name);
        }
        return newSelected;
    });
  }, []);

  const handleBulkToggle = useCallback(async (enable: boolean) => {
    const names = Array.from(selected);
    // Optimistic update
    setTools(prev => prev.map(t => names.includes(t.name) ? { ...t, disable: !enable } : t));

    try {
        await Promise.all(names.map(name => apiClient.setToolStatus(name, !enable)));
        toast({
            title: enable ? "Tools Enabled" : "Tools Disabled",
            description: `${names.length} tools have been ${enable ? "enabled" : "disabled"}.`
        });
        setSelected(new Set());
    } catch (e) {
        console.error("Failed to bulk toggle tools", e);
        fetchTools(); // Revert
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to update some tools."
        });
    }
  }, [selected, toast]);

  const isAllSelected = filteredTools.length > 0 && selected.size === filteredTools.length;

  // Reset selection when filtering changes
  useEffect(() => {
    setSelected(new Set());
  }, [searchQuery, selectedService, showPinnedOnly]);


  if (!isLoaded) {
      return (
          <div className="flex-1 p-8 animate-pulse text-muted-foreground">
              Loading tools...
          </div>
      );
  }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Tools</h2>
        <div className="flex items-center space-x-4">
            {selected.size > 0 && (
                <div className="flex items-center gap-2 animate-in fade-in slide-in-from-right-4 duration-300">
                    <span className="text-sm text-muted-foreground mr-2">{selected.size} selected</span>
                    <Button size="sm" variant="outline" onClick={() => handleBulkToggle(true)}>
                        <PlayCircle className="mr-2 h-4 w-4 text-green-600" /> Enable
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => handleBulkToggle(false)}>
                        <PauseCircle className="mr-2 h-4 w-4 text-amber-600" /> Disable
                    </Button>
                </div>
            )}
            <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search tools..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="w-[250px] pl-8 backdrop-blur-sm bg-background/50"
                />
            </div>
            <div className="flex items-center space-x-2">
                <Select value={selectedService} onValueChange={setSelectedService}>
                    <SelectTrigger className="w-[200px] backdrop-blur-sm bg-background/50">
                        <SelectValue placeholder="Filter by Service" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">All Services</SelectItem>
                        {services.map((service) => (
                            <SelectItem key={service.id} value={service.id}>
                                {service.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>
            <div className="flex items-center space-x-2">
                <Switch
                    id="show-pinned"
                    checked={showPinnedOnly}
                    onCheckedChange={setShowPinnedOnly}
                />
                <label htmlFor="show-pinned" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                    Show Pinned Only
                </label>
            </div>
            <Button
                variant="ghost"
                size="icon"
                onClick={toggleCompact}
                title={isCompact ? "Comfortable View" : "Compact View"}
                className="h-9 w-9"
            >
                {isCompact ? <LayoutList className="h-4 w-4" /> : <List className="h-4 w-4" />}
            </Button>
        </div>
      </div>

      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Available Tools</CardTitle>
          <CardDescription>Manage exposed tools from connected services.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[40px]">
                  <Checkbox
                    checked={isAllSelected}
                    onCheckedChange={(checked) => handleSelectAll(!!checked)}
                    aria-label="Select all"
                  />
                </TableHead>
                <TableHead className="w-[30px]"></TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Service</TableHead>
                <TableHead title="Estimated context size when tool is defined">Est. Context</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTools.map((tool) => (
                <TableRow key={tool.name} className={cn("group", isCompact ? "h-8" : "")}>
                  <TableCell className={isCompact ? "py-0 px-2" : ""}>
                    <Checkbox
                        checked={selected.has(tool.name)}
                        onCheckedChange={(checked) => handleSelectOne(tool.name, !!checked)}
                        aria-label={`Select ${tool.name}`}
                    />
                  </TableCell>
                  <TableCell className={isCompact ? "py-0 px-2" : ""}>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6"
                        onClick={() => togglePin(tool.name)}
                        aria-label={`Pin ${tool.name}`}
                      >
                          <Star className={`h-4 w-4 ${isPinned(tool.name) ? "fill-yellow-400 text-yellow-400" : "text-muted-foreground"}`} />
                      </Button>
                  </TableCell>
                  <TableCell className={cn("font-medium flex items-center", isCompact ? "py-0 px-2 h-8" : "")}>
                    <Wrench className={cn("mr-2 text-muted-foreground", isCompact ? "h-3 w-3" : "h-4 w-4")} />
                    {tool.name}
                  </TableCell>
                  <TableCell className={isCompact ? "py-0 px-2" : ""}>{tool.description}</TableCell>
                  <TableCell className={isCompact ? "py-0 px-2" : ""}>
                      <Badge variant="outline" className={isCompact ? "h-5 text-[10px] px-1" : ""}>{tool.serviceId}</Badge>
                  </TableCell>
                  <TableCell className={isCompact ? "py-0 px-2" : ""}>
                      <div className="flex items-center text-muted-foreground text-xs" title={`${estimateTokens(JSON.stringify(tool))} tokens`}>
                          <Info className="w-3 h-3 mr-1 opacity-50" />
                          {formatTokenCount(estimateTokens(JSON.stringify(tool)))}
                      </div>
                  </TableCell>
                  <TableCell className={isCompact ? "py-0 px-2" : ""}>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={!tool.disable}
                            onCheckedChange={() => toggleTool(tool.name, !tool.disable)}
                            className={isCompact ? "scale-75" : ""}
                        />
                        <span className={cn("text-muted-foreground", isCompact ? "text-[10px] w-12" : "text-sm w-16")}>
                            {!tool.disable ? "Enabled" : "Disabled"}
                        </span>
                    </div>
                  </TableCell>
                  <TableCell className={cn("text-right", isCompact ? "py-0 px-2" : "")}>
                      {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                      <Button variant="outline" size={isCompact ? "xs" as any : "sm"} onClick={() => openInspector(tool)} className={isCompact ? "h-6 px-2 text-[10px]" : ""}>
                          <Play className={cn("mr-1", isCompact ? "h-2 w-2" : "h-3 w-3")} /> Inspect
                      </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <ToolInspector
        tool={selectedTool}
        open={inspectorOpen}
        onOpenChange={setInspectorOpen}
      />
    </div>
  );
}
