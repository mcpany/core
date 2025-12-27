/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import React, { memo } from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import {
  Globe,
  Server,
  ShieldCheck,
  Activity,
  Database,
  Zap,
  Smartphone,
  Router,
  Lock,
  Wifi
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

// Generic Metric Row
const MetricRow = ({ label, value }: { label: string, value: string | number }) => (
  <div className="flex justify-between text-xs text-muted-foreground mt-1">
    <span>{label}</span>
    <span className="font-medium text-foreground">{value}</span>
  </div>
);

const NodeWrapper = ({ children, status, selected }: { children: React.ReactNode, status?: string, selected?: boolean }) => (
  <div className={`relative group transition-all duration-300 ${selected ? 'scale-105' : ''}`}>
    <div className={`
      absolute -inset-0.5 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-xl opacity-0 group-hover:opacity-50 blur transition duration-500
      ${selected ? 'opacity-75' : ''}
      ${status === 'warning' ? 'from-amber-500 to-orange-500' : ''}
      ${status === 'error' ? 'from-red-500 to-pink-500' : ''}
    `} />
    <Card className="relative min-w-[200px] border-0 shadow-lg bg-card/95 backdrop-blur-sm">
      {children}
    </Card>
  </div>
);

export const ClientNode = memo(({ data, selected }: NodeProps) => {
  return (
    <NodeWrapper selected={selected} status={data.status}>
      <CardContent className="p-4 flex flex-col items-center gap-2">
         {/* Input Handle */}
        <Handle type="source" position={Position.Bottom} className="!bg-primary/50 !w-3 !h-3" />

        <div className="p-3 bg-primary/10 rounded-full">
            <Smartphone className="w-6 h-6 text-primary" />
        </div>
        <div className="text-center">
            <h3 className="font-semibold text-sm">{data.label}</h3>
            <p className="text-xs text-muted-foreground">End User Device</p>
        </div>

        {data.metrics && (
             <div className="w-full mt-2 pt-2 border-t text-xs">
                <MetricRow label="Requests" value={data.metrics.requests} />
                <MetricRow label="Avg Size" value={data.metrics.avgReqSize} />
            </div>
        )}
      </CardContent>
    </NodeWrapper>
  );
});

export const GatewayNode = memo(({ data, selected }: NodeProps) => {
  return (
    <NodeWrapper selected={selected} status={data.status}>
      <Handle type="target" position={Position.Top} className="!bg-primary/50 !w-3 !h-3" />
      <CardContent className="p-4 flex flex-col items-center gap-2">
        <div className="p-3 bg-emerald-500/10 rounded-full relative">
             <ShieldCheck className="w-8 h-8 text-emerald-500" />
             <div className="absolute top-0 right-0 w-3 h-3 bg-emerald-500 rounded-full animate-pulse ring-2 ring-background" />
        </div>
        <div className="text-center">
            <h3 className="font-semibold text-sm">{data.label}</h3>
            <p className="text-xs text-muted-foreground">Firewall & MCP Router</p>
        </div>

        {data.metrics && (
             <div className="w-full mt-2 pt-2 border-t space-y-1">
                <div className="flex items-center justify-between text-xs">
                    <span className="text-emerald-500 flex items-center gap-1"><Lock className="w-3 h-3" /> Allowed</span>
                    <span>{data.metrics.allowed}</span>
                </div>
                 <div className="flex items-center justify-between text-xs">
                    <span className="text-red-500 flex items-center gap-1"><Lock className="w-3 h-3" /> Blocked</span>
                    <span>{data.metrics.blocked}</span>
                </div>
                 <MetricRow label="Cache Hit" value={data.metrics.cacheHitRate} />
            </div>
        )}
      </CardContent>
      <Handle type="source" position={Position.Bottom} className="!bg-primary/50 !w-3 !h-3" />
    </NodeWrapper>
  );
});

export const ServiceNode = memo(({ data, selected }: NodeProps) => {
   const isWarning = data.status === 'warning';
   const colorClass = isWarning ? 'text-amber-500' : 'text-blue-500';
   const bgClass = isWarning ? 'bg-amber-500/10' : 'bg-blue-500/10';

  return (
    <NodeWrapper selected={selected} status={data.status}>
      <Handle type="target" position={Position.Top} className="!bg-primary/50 !w-3 !h-3" />
      <CardContent className="p-3">
        <div className="flex items-start gap-3">
            <div className={`p-2 rounded-lg ${bgClass}`}>
                <Server className={`w-5 h-5 ${colorClass}`} />
            </div>
            <div>
                <h3 className="font-semibold text-sm">{data.label}</h3>
                <div className="flex items-center gap-2 mt-1">
                   <Badge variant={isWarning ? "destructive" : "secondary"} className="text-[10px] px-1 py-0 h-4">
                      {isWarning ? 'Issues' : 'Healthy'}
                   </Badge>
                   <span className="text-xs text-muted-foreground">{data.metrics?.latency} latency</span>
                </div>
            </div>
        </div>

         {data.metrics && (
             <div className="mt-3 pt-2 text-xs grid grid-cols-2 gap-2 border-t">
                <div>
                     <span className="text-muted-foreground block">Req/s</span>
                     <span className="font-medium">{data.metrics.requests}</span>
                </div>
                 <div>
                     <span className="text-muted-foreground block">Errors</span>
                     <span className={`font-medium ${data.metrics.errorRate !== '0%' ? 'text-red-500' : ''}`}>
                         {data.metrics.errorRate || '0%'}
                     </span>
                </div>
            </div>
        )}
      </CardContent>
      <Handle type="source" position={Position.Bottom} className="!bg-primary/50 !w-3 !h-3" />
    </NodeWrapper>
  );
});

ClientNode.displayName = 'ClientNode';
GatewayNode.displayName = 'GatewayNode';
ServiceNode.displayName = 'ServiceNode';
