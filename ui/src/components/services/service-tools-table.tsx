/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { ToolTable } from "@/components/tools/tool-table";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Search, Wrench } from "lucide-react";
import { ToolDefinition, apiClient } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { usePinnedTools } from "@/hooks/use-pinned-tools";
import { useToast } from "@/hooks/use-toast";
import { ToolInspector } from "@/components/tools/tool-inspector";

interface ServiceToolsTableProps {
  initialTools: ToolDefinition[];
  serviceId: string;
}

export function ServiceToolsTable({ initialTools, serviceId }: ServiceToolsTableProps) {
  const [tools, setTools] = useState<ToolDefinition[]>(initialTools);
  const [searchQuery, setSearchQuery] = useState("");
  const { isPinned, togglePin } = usePinnedTools();
  const { toast } = useToast();
  const [inspectorOpen, setInspectorOpen] = useState(false);
  const [selectedTool, setSelectedTool] = useState<ToolDefinition | null>(null);

  const filteredTools = tools.filter(tool =>
    tool.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    tool.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const toggleTool = async (name: string, currentStatus: boolean) => {
    // Optimistic update
    setTools(tools.map(t => t.name === name ? { ...t, disable: currentStatus } : t));

    try {
        await apiClient.setToolStatus(name, currentStatus); // currentStatus is the NEW status (disabled=true/false)
        toast({
            title: `Tool ${currentStatus ? "Disabled" : "Enabled"}`,
            description: `Tool ${name} has been ${currentStatus ? "disabled" : "enabled"}.`
        });
    } catch (e) {
        console.error("Failed to toggle tool", e);
        setTools(tools); // Revert? Ideally reload or revert specific item
        toast({
            variant: "destructive",
            title: "Failed to update tool status",
            description: String(e)
        });
    }
  };

  const openInspector = (tool: ToolDefinition) => {
      setSelectedTool(tool);
      setInspectorOpen(true);
  };

  if (!initialTools || initialTools.length === 0) {
     return (
       <Card>
        <CardHeader>
          <CardTitle className="text-xl flex items-center gap-2"><Wrench className="h-5 w-5" />Tools</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">No tools configured for this service.</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-xl flex items-center gap-2">
            <Wrench className="h-5 w-5" />
            Tools
            <Badge variant="secondary" className="ml-2">
                {tools.length}
            </Badge>
          </CardTitle>
          <div className="relative w-64">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search tools..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8 h-9"
            />
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="rounded-md border">
            <ToolTable
                tools={filteredTools}
                isCompact={false}
                isPinned={isPinned}
                togglePin={togglePin}
                toggleTool={toggleTool}
                openInspector={openInspector}
                // usageStats={} // We could fetch stats if we want, but keeping it simple for now
            />
             {filteredTools.length === 0 && (
                <div className="p-8 text-center text-muted-foreground">
                    No tools found matching "{searchQuery}"
                </div>
             )}
        </div>
      </CardContent>
    </Card>

    <ToolInspector
        tool={selectedTool}
        open={inspectorOpen}
        onOpenChange={setInspectorOpen}
    />
    </>
  );
}
