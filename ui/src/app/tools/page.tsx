/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, UpstreamServiceConfig, ToolAnalytics } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Wrench, Play, Star, Search, List, LayoutList, Layers } from "lucide-react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { ToolInspector } from "@/components/tools/tool-inspector";
import { usePinnedTools } from "@/hooks/use-pinned-tools";
import { estimateTokens, formatTokenCount } from "@/lib/tokens";
import { Info } from "lucide-react";
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
  const [stats, setStats] = useState<Record<string, ToolAnalytics>>({});
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
  }, []);

  const fetchTools = async () => {
    try {
      const [toolsRes, statsRes] = await Promise.all([
          apiClient.listTools(),
          apiClient.getToolUsage()
      ]);
      setTools(toolsRes?.tools || []);

      const statsMap: Record<string, ToolAnalytics> = {};
      if (statsRes) {
          statsRes.forEach(s => {
            // Use name@serviceId as key, similar to backend aggregation key logic but safer to match frontend tool objects
            statsMap[`${s.name}@${s.serviceId}`] = s;
          });
      }
      setStats(statsMap);
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

  const ToolTable = ({ tools }: { tools: ToolDefinition[] }) => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[30px]"></TableHead>
          <TableHead>Name</TableHead>
          <TableHead>Description</TableHead>
          <TableHead>Service</TableHead>
          <TableHead title="Estimated context size when tool is defined">Est. Context</TableHead>
          <TableHead title="Total executions">Calls</TableHead>
          <TableHead title="Success rate of executions">Success Rate</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {tools.map((tool) => (
          <TableRow key={tool.name} className={cn("group", isCompact ? "h-8" : "")}>
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
                <div className="text-xs font-mono">
                    {stats[`${tool.name}@${tool.serviceId}`]?.totalCalls || 0}
                </div>
            </TableCell>
            <TableCell className={isCompact ? "py-0 px-2" : ""}>
                 {(() => {
                     const s = stats[`${tool.name}@${tool.serviceId}`];
                     if (!s || s.totalCalls === 0) return <span className="text-muted-foreground text-xs">-</span>;
                     const successRate = 100 - s.failureRate;
                     let color = "text-green-500";
                     if (successRate < 70) color = "text-red-500";
                     else if (successRate < 90) color = "text-yellow-500";

                     return <span className={cn("font-mono text-xs", color)}>{successRate.toFixed(1)}%</span>;
                 })()}
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
                <Button variant="outline" size={isCompact ? "xs" as any : "sm"} onClick={() => openInspector(tool)} className={isCompact ? "h-6 px-2 text-[10px]" : ""}>
                    <Play className={cn("mr-1", isCompact ? "h-2 w-2" : "h-3 w-3")} /> Inspect
                </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );

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
            <ToolTable tools={filteredTools} />
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
                    <ToolTable tools={groupTools} />
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
