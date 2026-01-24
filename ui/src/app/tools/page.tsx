/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, UpstreamServiceConfig, ToolAnalytics } from "@/lib/client";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Search, List, LayoutList, Layers } from "lucide-react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { ToolInspector } from "@/components/tools/tool-inspector";
import { usePinnedTools } from "@/hooks/use-pinned-tools";
import { ToolTable } from "@/components/tools/tool-table";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";

/**
 * ToolsPage component.
 * @returns The rendered component.
 */
export default function ToolsPage() {
  const [tools, setTools] = useState<ToolDefinition[]>([]);
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [toolUsage, setToolUsage] = useState<Record<string, ToolAnalytics>>({});
  const [selectedTool, setSelectedTool] = useState<ToolDefinition | null>(null);
  const [inspectorOpen, setInspectorOpen] = useState(false);
  const { isPinned, togglePin, isLoaded } = usePinnedTools();
  const [showPinnedOnly, setShowPinnedOnly] = useState(false);
  const [selectedService, setSelectedService] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [isCompact, setIsCompact] = useState(false);
  const [groupBy, setGroupBy] = useState<"none" | "service" | "category">("none");

  useEffect(() => {
    const savedCompact = localStorage.getItem("tools_compact_view") === "true";
    setIsCompact(savedCompact);
  }, []);

  useEffect(() => {
    fetchTools();
    fetchServices();
    fetchToolUsage();
  }, []);

  const fetchToolUsage = async () => {
    try {
        const stats = await apiClient.getToolUsage();
        const statsMap: Record<string, ToolAnalytics> = {};
        stats.forEach(s => {
            statsMap[`${s.name}@${s.serviceId}`] = s;
        });
        setToolUsage(statsMap);
    } catch (e) {
        console.error("Failed to fetch tool usage", e);
    }
  };

  const fetchTools = async () => {
    try {
      const res = await apiClient.listTools();
      setTools(res?.tools || []);
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

  // Grouping logic
  const groupedTools = filteredTools.reduce((acc, tool) => {
    let key = "Other";
    if (groupBy === "service") {
      const service = services.find((s) => s.id === tool.serviceId);
      key = service ? service.name : tool.serviceId || "Unknown Service";
    } else if (groupBy === "category") {
      key = tool.tags && tool.tags.length > 0 ? tool.tags[0] : "Uncategorized";
    }

    if (!acc[key]) {
      acc[key] = [];
    }
    acc[key].push(tool);
    return acc;
  }, {} as Record<string, ToolDefinition[]>);

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
                <Select value={groupBy} onValueChange={(v: any) => setGroupBy(v)}>
                    <SelectTrigger className="w-[180px] backdrop-blur-sm bg-background/50">
                        <Layers className="mr-2 h-4 w-4" />
                        <SelectValue placeholder="Group By" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="none">No Grouping</SelectItem>
                        <SelectItem value="service">Group by Service</SelectItem>
                        <SelectItem value="category">Group by Category</SelectItem>
                    </SelectContent>
                </Select>
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
          {groupBy === "none" ? (
            <ToolTable
              tools={filteredTools}
              isCompact={isCompact}
              isPinned={isPinned}
              togglePin={togglePin}
              toggleTool={toggleTool}
              openInspector={openInspector}
              usageStats={toolUsage}
            />
          ) : (
            <Accordion type="multiple" defaultValue={Object.keys(groupedTools)} className="w-full">
              {Object.entries(groupedTools).map(([groupName, groupTools]) => (
                <AccordionItem key={groupName} value={groupName}>
                  <AccordionTrigger className="hover:no-underline px-2">
                    <span className="font-medium text-lg flex items-center">
                      {groupName}
                      <Badge variant="secondary" className="ml-2 text-xs">
                        {groupTools.length}
                      </Badge>
                    </span>
                  </AccordionTrigger>
                  <AccordionContent>
                    <ToolTable
                      tools={groupTools}
                      isCompact={isCompact}
                      isPinned={isPinned}
                      togglePin={togglePin}
                      toggleTool={toggleTool}
                      openInspector={openInspector}
                      usageStats={toolUsage}
                    />
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          )}
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
