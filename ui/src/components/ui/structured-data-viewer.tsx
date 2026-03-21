/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { cn } from "@/lib/utils";
import { Check, Copy } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface StructuredDataViewerProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: any;
  maxHeight?: number;
  className?: string;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const renderValue = (value: any, isRoot = false): React.ReactNode => {
  if (value === null)
    return <span className="text-muted-foreground italic">null</span>;
  if (value === undefined)
    return <span className="text-muted-foreground italic">undefined</span>;

  const type = typeof value;

  if (type === "string") {
    return (
      <span className="text-green-600 dark:text-green-400 break-all">
        "{value}"
      </span>
    );
  }
  if (type === "number") {
    return <span className="text-blue-600 dark:text-blue-400">{value}</span>;
  }
  if (type === "boolean") {
    return (
      <span className="text-amber-600 dark:text-amber-400 font-medium">
        {value ? "true" : "false"}
      </span>
    );
  }
  if (Array.isArray(value)) {
    if (value.length === 0)
      return <span className="text-muted-foreground">[]</span>;
    return (
      <div className="pl-4 border-l border-border/50 ml-2 space-y-1">
        {value.map((item, index) => (
          <div key={index} className="flex gap-2">
            <span className="text-muted-foreground select-none">-</span>
            <div className="flex-1">{renderValue(item)}</div>
          </div>
        ))}
      </div>
    );
  }
  if (type === "object") {
    const keys = Object.keys(value);
    if (keys.length === 0)
      return <span className="text-muted-foreground">{"{}"}</span>;

    return (
      <div
        className={cn(
          "grid gap-[1px] bg-border/20",
          isRoot ? "" : "pl-4 border-l border-border/50 ml-2 mt-1",
        )}
      >
        {keys.map((key) => (
          <div
            key={key}
            className="bg-background/50 backdrop-blur-sm grid grid-cols-1 md:grid-cols-[1fr_2fr] gap-4 p-2 transition-colors hover:bg-muted/30"
          >
            <div className="text-sm font-medium text-foreground/80 break-words">
              {key}
            </div>
            <div className="text-sm">{renderValue(value[key])}</div>
          </div>
        ))}
      </div>
    );
  }
  return <span>{String(value)}</span>;
};

export function StructuredDataViewer({
  data,
  maxHeight,
  className,
}: StructuredDataViewerProps) {
  const { toast } = useToast();
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(JSON.stringify(data, null, 2));
    setCopied(true);
    toast({ title: "Copied payload to clipboard" });
    setTimeout(() => setCopied(false), 2000);
  };

  if (!data) {
    return (
      <div className="p-4 text-sm text-muted-foreground italic bg-muted/10 rounded-md border border-border/50">
        No data available
      </div>
    );
  }

  return (
    <div
      className={cn(
        "relative group rounded-md border border-border/50 overflow-hidden bg-muted/5",
        className,
      )}
    >
      <div className="absolute right-2 top-2 z-10 opacity-0 group-hover:opacity-100 transition-opacity">
        <button
          onClick={handleCopy}
          className="p-1.5 rounded-md bg-background/80 backdrop-blur border border-border/50 text-muted-foreground hover:text-foreground shadow-sm transition-colors"
        >
          {copied ? (
            <Check className="h-3 w-3 text-green-500" />
          ) : (
            <Copy className="h-3 w-3" />
          )}
        </button>
      </div>
      <div
        className="overflow-auto"
        style={maxHeight ? { maxHeight: `${maxHeight}px` } : undefined}
      >
        {renderValue(data, true)}
      </div>
    </div>
  );
}
