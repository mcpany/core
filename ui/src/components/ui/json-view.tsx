/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { cn } from "@/lib/utils";
import { Copy, Check } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { JsonTree } from "./json-tree";

interface JsonViewProps {
  data: unknown;
  className?: string;
  defaultExpandDepth?: number;
}

/**
 * JsonView component.
 * @param props - The component props.
 * @param props.data - The data to display.
 * @param props.className - The className.
 * @param props.defaultExpandDepth - The default expansion depth.
 * @returns The rendered component.
 */
export function JsonView({ data, className, defaultExpandDepth = 1 }: JsonViewProps) {
  const [copied, setCopied] = useState(false);
  const { toast } = useToast();

  const handleCopy = () => {
    const text = JSON.stringify(data, null, 2);
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).catch(e => console.error("Clipboard error", e));
    }
    setCopied(true);
    toast({
        title: "Copied",
        description: "JSON copied to clipboard",
    });
    setTimeout(() => setCopied(false), 2000);
  };

  if (data === undefined || data === null) {
      return <span className="text-muted-foreground italic">null</span>;
  }

  return (
    <div className={cn("relative group bg-muted/20 border rounded-md overflow-hidden", className)}>
        <Button
            variant="ghost"
            size="icon"
            className="absolute right-2 top-2 h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity bg-background/50 hover:bg-background z-10"
            onClick={handleCopy}
            title="Copy JSON"
        >
            {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
        </Button>
        <div className="p-3 text-xs overflow-auto max-h-[inherit]">
             <JsonTree data={data} defaultExpandDepth={defaultExpandDepth} />
        </div>
    </div>
  );
}
