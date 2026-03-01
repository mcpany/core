/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from "react";
import { ChevronDown, ChevronRight, Copy, Check } from "lucide-react";
import { cn } from "@/lib/utils";

interface PropertiesTableProps {
  data: Record<string, any> | undefined;
  className?: string;
}

interface PropertyRowProps {
  propKey: string;
  value: any;
  depth: number;
}

const PropertyRow: React.FC<PropertyRowProps> = ({ propKey, value, depth }) => {
  const [expanded, setExpanded] = useState(false);
  const [copied, setCopied] = useState(false);

  const isObject = value !== null && typeof value === "object";
  const isArray = Array.isArray(value);

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigator.clipboard.writeText(
      typeof value === "string" ? value : JSON.stringify(value, null, 2)
    );
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const renderValue = () => {
    if (value === null) return <span className="text-muted-foreground italic">null</span>;
    if (value === undefined) return <span className="text-muted-foreground italic">undefined</span>;
    if (typeof value === "boolean") {
      return <span className={value ? "text-green-500" : "text-red-500"}>{value.toString()}</span>;
    }
    if (typeof value === "number") {
      return <span className="text-blue-500 dark:text-blue-400">{value}</span>;
    }
    if (typeof value === "string") {
      // Truncate very long strings for initial view, maybe allow expanding?
      // For now, we'll let it wrap.
      return <span className="text-amber-600 dark:text-amber-400 break-all">"{value}"</span>;
    }
    if (isArray) {
      return <span className="text-muted-foreground italic">Array({value.length})</span>;
    }
    if (isObject) {
      return <span className="text-muted-foreground italic">Object({Object.keys(value).length} keys)</span>;
    }
    return String(value);
  };

  return (
    <>
      <div
        className={cn(
          "flex items-start py-2 px-3 border-b border-border/40 hover:bg-muted/30 transition-colors group",
          isObject && "cursor-pointer"
        )}
        onClick={() => isObject && setExpanded(!expanded)}
      >
        <div
          className="flex-1 flex items-center min-w-[200px]"
          style={{ paddingLeft: `${depth * 16}px` }}
        >
          <div className="w-4 h-4 mr-1 flex items-center justify-center">
            {isObject && (
              expanded ? <ChevronDown className="h-3 w-3 text-muted-foreground" /> : <ChevronRight className="h-3 w-3 text-muted-foreground" />
            )}
          </div>
          <span className="font-mono text-xs font-medium text-foreground">{propKey}</span>
        </div>

        <div className="flex-[2] flex items-center justify-between pl-4 border-l border-border/30 overflow-hidden group/value">
          <div className="font-mono text-xs truncate mr-2 w-full pr-4">
            {renderValue()}
          </div>
          <button
            onClick={handleCopy}
            className="opacity-0 group-hover/value:opacity-100 transition-opacity p-1 hover:bg-muted rounded text-muted-foreground"
            title="Copy value"
          >
            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
          </button>
        </div>
      </div>

      {expanded && isObject && (
        <div className="w-full">
          {Object.entries(value).map(([k, v]) => (
            <PropertyRow key={k} propKey={k} value={v} depth={depth + 1} />
          ))}
        </div>
      )}
    </>
  );
};

/**
 * PropertiesTable component to display key-value pairs of objects.
 * @param props The component props.
 * @returns The rendered PropertiesTable.
 */
export function PropertiesTable({ data, className }: PropertiesTableProps) {
  if (!data || Object.keys(data).length === 0) {
    return (
      <div className={cn("flex items-center justify-center p-8 text-sm text-muted-foreground bg-muted/10 rounded-md border border-dashed", className)}>
        No properties
      </div>
    );
  }

  return (
    <div className={cn("w-full border rounded-md bg-background overflow-hidden", className)}>
      <div className="flex text-xs font-medium text-muted-foreground border-b p-2 bg-muted/20">
        <div className="flex-1 pl-8">Property</div>
        <div className="flex-[2] pl-4 border-l">Value</div>
      </div>
      <div className="flex flex-col w-full">
        {Object.entries(data).map(([key, value]) => (
          <PropertyRow key={key} propKey={key} value={value} depth={0} />
        ))}
      </div>
    </div>
  );
}
