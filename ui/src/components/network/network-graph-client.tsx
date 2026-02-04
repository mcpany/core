/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useMemo, useCallback } from 'react';
import { useRouter } from "next/navigation";
import {
  ReactFlow,

  Controls,
  Background,
  BackgroundVariant,
  Panel,
  Node,
  ReactFlowProvider,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetDescription,
} from "@/components/ui/sheet";
import {
    RefreshCcw,
    Zap,
    Server,
    Database,
    Users,
    Activity,
    Webhook,
    Layers,
    MessageSquare,
    Link as LinkIcon,
    AlertTriangle,
    CheckCircle2,
    XCircle,
    Cpu,
    Filter,
    ChevronDown,
    ChevronRight,
} from "lucide-react";

import { NetworkLegend } from "@/components/network/network-legend";
import { useNetworkTopology } from "@/hooks/use-network-topology";
import { useIsMobile } from "@/hooks/use-mobile";
import { NodeType, NodeStatus } from "@/types/topology";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface NodeData extends Record<string, unknown> {
    label: string;
    type: NodeType;
    status: NodeStatus;
    metrics?: { qps?: number; latencyMs?: number; errorRate?: number };
    metadata?: Record<string, string>;
}

/**
 * NodeIcon component.
 * @param props - The component props.
 * @param props.type - The type definition.
 * @returns The rendered component.
 */
const NodeIcon = React.memo(({ type }: { type: NodeType }) => {
    switch (type) {
        case 'NODE_TYPE_CORE': return <Cpu className="h-4 w-4 text-blue-500 dark:text-blue-400" />;
        case 'NODE_TYPE_SERVICE': return <Server className="h-4 w-4 text-indigo-500 dark:text-indigo-400" />;
        case 'NODE_TYPE_CLIENT': return <Users className="h-4 w-4 text-green-500 dark:text-green-400" />;
        case 'NODE_TYPE_TOOL': return <Zap className="h-4 w-4 text-amber-500 dark:text-amber-400" />;
        case 'NODE_TYPE_API_CALL': return <LinkIcon className="h-4 w-4 text-slate-500 dark:text-slate-400" />;
        case 'NODE_TYPE_MIDDLEWARE': return <Layers className="h-4 w-4 text-orange-500 dark:text-orange-400" />;
        case 'NODE_TYPE_WEBHOOK': return <Webhook className="h-4 w-4 text-pink-500 dark:text-pink-400" />;
        case 'NODE_TYPE_RESOURCE': return <Database className="h-4 w-4 text-cyan-500 dark:text-cyan-400" />;
        case 'NODE_TYPE_PROMPT': return <MessageSquare className="h-4 w-4 text-purple-500 dark:text-purple-400" />;
        default: return <Activity className="h-4 w-4 text-gray-500 dark:text-gray-400" />;
    }
});
NodeIcon.displayName = 'NodeIcon';

/**
 * StatusIcon component.
 * @param props - The component props.
 * @param props.status - The current status.
 * @returns The rendered component.
 */
const StatusIcon = React.memo(({ status }: { status: NodeStatus }) => {
    switch (status) {
        case 'NODE_STATUS_ACTIVE': return <CheckCircle2 className="h-4 w-4 text-green-500" />;
        case 'NODE_STATUS_INACTIVE': return <XCircle className="h-4 w-4 text-slate-400" />;
        case 'NODE_STATUS_ERROR': return <AlertTriangle className="h-4 w-4 text-destructive" />;
        default: return <div className="h-4 w-4 rounded-full border-2 border-slate-300" />;
    }
});
StatusIcon.displayName = 'StatusIcon';

// Optimization: Memoize defaultEdgeOptions to prevent re-renders in ReactFlow
const defaultEdgeOptions = {
    type: 'smoothstep',
    animated: true,
    style: { strokeWidth: 2 }
};

/**
 * Props for the NetworkGraphFlow component.
 */
export interface NetworkGraphFlowProps {
    /**
     * Whether to render in widget mode (simplified UI).
     */
    widgetMode?: boolean;
}

/**
 * NetworkGraphFlow component.
 * Renders the interactive network graph using ReactFlow.
 * @param props - The component props.
 * @returns The rendered component.
 */
export function NetworkGraphFlow({ widgetMode = false }: NetworkGraphFlowProps) {
  const router = useRouter();
  const { nodes, edges, onNodesChange, onEdgesChange, onConnect, refreshTopology, autoLayout } = useNetworkTopology();
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const isMobile = useIsMobile();

  // ⚡ BOLT: Derive selected node from current nodes list using ID to ensure live data updates.
  // Randomized Selection from Top 5 High-Impact Targets
  const selectedNode = useMemo(() => {
    if (!selectedNodeId) return null;
    return nodes.find((n) => n.id === selectedNodeId) || null;
  }, [nodes, selectedNodeId]);
  const [isControlsExpanded, setIsControlsExpanded] = useState(true);

  // Basic filtering state
  const [showSystem, setShowSystem] = useState(true);
  const [showTools, setShowTools] = useState(!widgetMode); // Hide detailed tools by default in widget mode
  const [showLegend, setShowLegend] = useState(false);

  // Collapse controls by default on mobile
  React.useEffect(() => {
    if (isMobile) {
        setIsControlsExpanded(false);
    } else {
        setIsControlsExpanded(true);
    }
  }, [isMobile]);

  const filteredNodes = useMemo(() => {
    return nodes.filter(n => {
        const type = n.data.type as NodeType;
        if (!showSystem && (type === 'NODE_TYPE_MIDDLEWARE' || type === 'NODE_TYPE_CORE')) return false;
        if (!showTools && (type === 'NODE_TYPE_TOOL' || type === 'NODE_TYPE_RESOURCE' || type === 'NODE_TYPE_PROMPT')) return false;
        return true;
    });
  }, [nodes, showSystem, showTools]);

  const filteredEdges = useMemo(() => {
      const nodeIds = new Set(filteredNodes.map(n => n.id));
      return edges.filter(e => nodeIds.has(e.source) && nodeIds.has(e.target));
  }, [edges, filteredNodes]);

  // Optimization: Memoize onNodeClick to prevent unnecessary re-renders in ReactFlow
  const onNodeClick = useCallback((event: React.MouseEvent, node: Node) => {
      setSelectedNodeId(node.id);
      setIsSheetOpen(true);
  }, []);

  const getStatusBadgeVariant = (status: NodeStatus) => {
      switch (status) {
          case 'NODE_STATUS_ACTIVE': return 'default'; // primary color usually indicates active/good
          case 'NODE_STATUS_INACTIVE': return 'secondary';
          case 'NODE_STATUS_ERROR': return 'destructive';
          default: return 'outline';
      }
  };

  return (
    <div className={cn("h-full w-full relative bg-muted/5 group", widgetMode && "rounded-lg overflow-hidden")}>
      {!widgetMode && (
          <div className={cn("absolute top-4 left-4 right-4 z-10 space-y-4 pointer-events-none flex flex-col gap-2 transition-all", isMobile ? "items-start" : "")}>
              <Card className={cn(
                  "pointer-events-auto backdrop-blur-md bg-background/80 shadow-lg border-muted/60 transition-all hover:shadow-xl overflow-hidden",
                  isControlsExpanded ? "w-[320px]" : "w-auto"
              )}>
                  <CardHeader className="p-4 pb-2 cursor-pointer" onClick={() => setIsControlsExpanded(!isControlsExpanded)}>
                      <CardTitle className="text-lg flex items-center justify-between gap-4">
                          <div className="flex items-center gap-2">
                            <Activity className="h-5 w-5 text-primary" />
                            <span className={cn(isControlsExpanded ? "block" : "hidden sm:block")}>Network Graph</span>
                          </div>
                          <div className="flex items-center gap-2">
                            <Badge variant="outline" className="font-normal text-xs">{filteredNodes.length} Nodes</Badge>
                            {isMobile && (
                                <Button variant="ghost" size="icon" className="h-6 w-6">
                                    <Filter className="h-3 w-3" />
                                </Button>
                            )}
                          </div>
                      </CardTitle>
                      {isControlsExpanded && (
                        <CardDescription className="text-xs">
                            Live topology of MCP services and tools.
                        </CardDescription>
                      )}
                  </CardHeader>
                  {isControlsExpanded && (
                      <>
                        <CardContent className="p-4 pt-2 flex gap-2">
                            <Button variant="outline" size="sm" onClick={refreshTopology} className="flex-1 h-8 text-xs">
                                <RefreshCcw className="mr-2 h-3 w-3" /> Refresh
                            </Button>
                            <Button size="sm" onClick={autoLayout} className="flex-1 h-8 text-xs bg-primary/90 hover:bg-primary">
                                <Zap className="mr-2 h-3 w-3" /> Layout
                            </Button>
                        </CardContent>
                        <div className="px-4 pb-4">
                            <div className="flex items-center justify-between">
                                <span className="text-xs text-muted-foreground font-medium flex items-center gap-1">
                                    <Filter className="h-3 w-3" /> Filters
                                </span>
                            </div>
                            <div className="mt-2 space-y-2">
                                <div className="flex items-center space-x-2">
                                    <Checkbox id="show-system" checked={showSystem} onCheckedChange={(c) => setShowSystem(!!c)} />
                                    <Label htmlFor="show-system" className="text-xs font-normal cursor-pointer">Show System (Core/Middleware)</Label>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <Checkbox id="show-tools" checked={showTools} onCheckedChange={(c) => setShowTools(!!c)} />
                                    <Label htmlFor="show-tools" className="text-xs font-normal cursor-pointer">Show Capability Details (Tools)</Label>
                                </div>

                                <div className="pt-2 mt-2 border-t">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => setShowLegend(!showLegend)}
                                        className="w-full justify-start h-6 px-0 text-xs text-muted-foreground hover:text-foreground"
                                    >
                                        {showLegend ? <ChevronDown className="mr-2 h-3 w-3" /> : <ChevronRight className="mr-2 h-3 w-3" />}
                                        Show Legend
                                    </Button>
                                    {showLegend && (
                                        <div className="mt-2 pl-1">
                                            <NetworkLegend />
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                      </>
                  )}
              </Card>
          </div>
      )}

      <ReactFlow
        nodes={filteredNodes}
        edges={filteredEdges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onNodeClick={onNodeClick}
        fitView
        attributionPosition="bottom-right"
        className="bg-background"
        minZoom={widgetMode ? 0.2 : 0.1}
        maxZoom={widgetMode ? 1.0 : 1.5}
        defaultEdgeOptions={defaultEdgeOptions}
        nodesDraggable={!widgetMode}
        panOnDrag={!widgetMode}
        zoomOnScroll={!widgetMode}
        preventScrolling={widgetMode}
      >
        {!widgetMode && (
            <Controls showInteractive={false} className="bg-background/80 backdrop-blur border-muted shadow-sm dark:bg-slate-900/80 dark:border-slate-800 dark:text-slate-200 [&>button]:!border-muted [&>button]:!bg-transparent hover:[&>button]:!bg-muted" />
        )}

        <Background variant={BackgroundVariant.Dots} gap={24} size={1} color="currentColor" className="text-muted-foreground/20" />

        {!widgetMode && (
            <Panel position="bottom-center" className="mb-8">
                <div className="bg-background/90 p-2 px-4 rounded-full border shadow-lg backdrop-blur text-[10px] text-muted-foreground flex gap-4 items-center">
                    <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-blue-500"></div> Core</div>
                    <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-indigo-500"></div> Service</div>
                    <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-green-500"></div> Client</div>
                    <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-amber-500"></div> Tool</div>
                    <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-orange-500"></div> Middleware</div>
                    <div className="flex items-center gap-1.5"><div className="w-2 h-2 rounded-full bg-red-500"></div> Error</div>
                </div>
            </Panel>
        )}
      </ReactFlow>

      <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
        <SheetContent className="w-full sm:w-[540px] border-l-muted">
            <SheetHeader className="pb-4 border-b">
                <div className="flex items-center gap-2 text-muted-foreground text-sm uppercase tracking-wider font-mono mb-1">
                    {selectedNode?.data.type.replace('NODE_TYPE_', '')}
                </div>
                <SheetTitle className="flex items-center gap-3 text-2xl">
                    {selectedNode && <NodeIcon type={selectedNode.data.type} />}
                    {selectedNode?.data.label}
                </SheetTitle>
                <SheetDescription className="flex items-center gap-2">
                    ID: <code className="bg-muted px-1 py-0.5 rounded text-xs font-mono">{selectedNode?.id}</code>
                </SheetDescription>
            </SheetHeader>

            <div className="py-6 space-y-6">
                {/* Status Section */}
                <div className="flex items-center justify-between p-4 bg-muted/30 rounded-lg border">
                    <div className="flex flex-col">
                        <span className="text-sm font-medium">Operational Status</span>
                        <span className="text-xs text-muted-foreground">Current health check result</span>
                    </div>
                    <Badge variant={selectedNode ? getStatusBadgeVariant(selectedNode.data.status) : 'outline'} className="flex items-center gap-1.5 px-3 py-1">
                        {selectedNode && <StatusIcon status={selectedNode.data.status} />}
                        {selectedNode?.data.status.replace('NODE_STATUS_', '')}
                    </Badge>
                </div>

                {/* Metrics Section */}
                {selectedNode?.data.metrics && (
                    <div className="space-y-3">
                        <h4 className="font-medium text-sm flex items-center gap-2">
                            <Activity className="h-4 w-4" /> Real-time Metrics
                        </h4>
                        <div className="grid grid-cols-3 gap-3">
                            <MetricCard label="QPS" value={selectedNode.data.metrics.qps?.toFixed(2)} unit="req/s" />
                            <MetricCard label="Latency" value={selectedNode.data.metrics.latencyMs?.toFixed(0)} unit="ms" />
                            <MetricCard
                                label="Error Rate"
                                value={((selectedNode.data.metrics.errorRate || 0) * 100).toFixed(2)}
                                unit="%"
                                intent={((selectedNode.data.metrics.errorRate || 0) > 0.05) ? "danger" : "neutral"}
                            />
                        </div>
                    </div>
                )}

                 {/* Metadata Section */}
                 {selectedNode?.data.metadata && Object.keys(selectedNode.data.metadata).length > 0 && (
                    <div className="space-y-3">
                        <h4 className="font-medium text-sm flex items-center gap-2">
                            <Database className="h-4 w-4" /> Metadata
                        </h4>
                        <div className="bg-black/90 rounded-lg p-3 overflow-hidden border border-slate-800">
                             <table className="w-full text-xs font-mono">
                                <tbody>
                                    {Object.entries(selectedNode.data.metadata).map(([k, v]) => (
                                        <tr key={k} className="border-b border-white/10 last:border-0">
                                            <td className="py-2 pr-4 text-slate-400 select-none w-1/3">{k}</td>
                                            <td className="py-2 text-slate-100 break-all">{String(v)}</td>
                                        </tr>
                                    ))}
                                </tbody>
                             </table>
                        </div>
                    </div>
                )}

                {/* Actions */}
                <div className="pt-4">
                     <h4 className="font-medium text-sm mb-3">Quick Actions</h4>
                     <div className="flex flex-wrap gap-2">
                         <Button
                            variant="outline"
                            size="sm"
                            className="h-8"
                            onClick={() => {
                                if (!selectedNode) return;
                                const id = selectedNode.id;
                                let source = "ALL";
                                // Parse ID to determine source
                                if (id === "mcp-core") source = "core";
                                else if (id.startsWith("svc-")) source = id.replace("svc-", "");
                                // For tools or others, we default to ALL or keep generic

                                router.push(`/logs?source=${source}`);
                            }}
                         >
                            View Logs
                         </Button>
                         <Button variant="outline" size="sm" className="h-8">Trace Request</Button>
                         {selectedNode?.data.status === 'NODE_STATUS_ERROR' && (
                             <Button variant="destructive" size="sm" className="h-8">Restart Service</Button>
                         )}
                     </div>
                </div>
            </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}

/**
 * MetricCard component.
 * @param props - The component props.
 * @param props.label - The label property.
 * @param props.value - The current value.
 * @param props.unit - The unit property.
 * @param props.intent - The intent property.
 * @returns The rendered component.
 */
const MetricCard = React.memo(({ label, value, unit, intent = "neutral" }: { label: string, value?: string, unit: string, intent?: "neutral" | "danger" | "success" }) => {
    return (
        <Card className={`p-3 bg-card/50 ${intent === 'danger' ? 'border-red-200 bg-red-50 dark:bg-red-900/10 dark:border-red-900' : ''}`}>
            <div className="text-[10px] text-muted-foreground uppercase tracking-wider font-semibold">{label}</div>
            <div className={`text-xl font-bold mt-1 ${intent === 'danger' ? 'text-red-600 dark:text-red-400' : ''}`}>
                {value || "-"} <span className="text-xs font-normal text-muted-foreground">{unit}</span>
            </div>
        </Card>
    )
});
MetricCard.displayName = 'MetricCard';

/**
 * NetworkGraphClient component.
 * @returns The rendered component.
 */
export function NetworkGraphClient() {
    return (
        <ReactFlowProvider>
            <NetworkGraphFlow />
        </ReactFlowProvider>
    )
}
