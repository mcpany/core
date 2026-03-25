/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { User, Cpu, Terminal, Database, Globe } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';

const NodeWrapper = ({ children, className, selected }: { children: React.ReactNode, className?: string, selected?: boolean }) => (
  <div className={cn(
    "px-4 py-2 shadow-md rounded-md bg-background border-2 min-w-[150px]",
    selected ? "border-primary" : "border-muted-foreground/20",
    className
  )}>
    {children}
  </div>
);

/**
 * UserNode represents a user in the flow.
 */
export const UserNode = memo(({ data, selected }: any) => {
  return (
    <NodeWrapper className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800" selected={selected}>
      <Handle type="source" position={Position.Bottom} className="w-3 h-3 !bg-blue-500" />
      <div className="flex items-center gap-2">
        <div className="p-2 rounded-full bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300">
            <User size={16} />
        </div>
        <div className="flex flex-col">
            <span className="text-sm font-bold">{data.label}</span>
            <span className="text-[10px] text-muted-foreground">Human</span>
        </div>
      </div>
    </NodeWrapper>
  );
});
UserNode.displayName = 'UserNode';

/**
 * AgentNode represents an agent in the flow.
 */
export const AgentNode = memo(({ data, selected }: any) => {
  return (
    <NodeWrapper className="bg-purple-50 dark:bg-purple-950/20 border-purple-200 dark:border-purple-800" selected={selected}>
      <Handle type="target" position={Position.Top} className="w-3 h-3 !bg-purple-500" />
      <Handle type="source" position={Position.Bottom} className="w-3 h-3 !bg-purple-500" />
      <div className="flex items-center gap-2">
        <div className="p-2 rounded-full bg-purple-100 dark:bg-purple-900 text-purple-600 dark:text-purple-300">
            <Cpu size={16} />
        </div>
        <div className="flex flex-col">
            <span className="text-sm font-bold">{data.label}</span>
            <span className="text-[10px] text-muted-foreground">{data.role || 'Agent'}</span>
        </div>
      </div>
      {data.status && (
          <Badge variant="outline" className="mt-2 text-[10px] w-full justify-center bg-background/50">
              {data.status}
          </Badge>
      )}
    </NodeWrapper>
  );
});
AgentNode.displayName = 'AgentNode';

/**
 * ToolNode represents a tool in the flow.
 */
export const ToolNode = memo(({ data, selected }: any) => {
  return (
    <NodeWrapper className="bg-amber-50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-800" selected={selected}>
      <Handle type="target" position={Position.Top} className="w-3 h-3 !bg-amber-500" />
      <div className="flex items-center gap-2">
        <div className="p-2 rounded-full bg-amber-100 dark:bg-amber-900 text-amber-600 dark:text-amber-300">
            <Terminal size={16} />
        </div>
        <div className="flex flex-col">
            <span className="text-sm font-bold">{data.label}</span>
            <span className="text-[10px] text-muted-foreground">Tool</span>
        </div>
      </div>
    </NodeWrapper>
  );
});
ToolNode.displayName = 'ToolNode';

/**
 * ResourceNode represents a resource in the flow.
 */
export const ResourceNode = memo(({ data, selected }: any) => {
    return (
      <NodeWrapper className="bg-cyan-50 dark:bg-cyan-950/20 border-cyan-200 dark:border-cyan-800" selected={selected}>
        <Handle type="target" position={Position.Top} className="w-3 h-3 !bg-cyan-500" />
        <Handle type="source" position={Position.Bottom} className="w-3 h-3 !bg-cyan-500" />
        <div className="flex items-center gap-2">
          <div className="p-2 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-600 dark:text-cyan-300">
              <Database size={16} />
          </div>
          <div className="flex flex-col">
              <span className="text-sm font-bold">{data.label}</span>
              <span className="text-[10px] text-muted-foreground">Resource</span>
          </div>
        </div>
      </NodeWrapper>
    );
  });
ResourceNode.displayName = 'ResourceNode';

/**
 * ServiceNode represents a service in the flow.
 */
export const ServiceNode = memo(({ data, selected }: any) => {
    return (
      <NodeWrapper className="bg-indigo-50 dark:bg-indigo-950/20 border-indigo-200 dark:border-indigo-800" selected={selected}>
        <Handle type="target" position={Position.Top} className="w-3 h-3 !bg-indigo-500" />
        <Handle type="source" position={Position.Bottom} className="w-3 h-3 !bg-indigo-500" />
        <div className="flex items-center gap-2">
          <div className="p-2 rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-300">
              <Globe size={16} />
          </div>
          <div className="flex flex-col">
              <span className="text-sm font-bold">{data.label}</span>
              <span className="text-[10px] text-muted-foreground">Service</span>
          </div>
        </div>
      </NodeWrapper>
    );
  });
ServiceNode.displayName = 'ServiceNode';
