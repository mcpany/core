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

interface JsonViewProps {
  data: unknown;
  className?: string;
}

/**
 * JsonView component.
 * @param props - The component props.
 * @param props.data - The data to display.
 * @param props.className - The className.
 * @returns The rendered component.
 */
export function JsonView({ data, className }: JsonViewProps) {
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
    <div className={cn("relative group", className)}>
        <Button
            variant="ghost"
            size="icon"
            className="absolute right-2 top-2 h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity bg-background/50 hover:bg-background"
            onClick={handleCopy}
            title="Copy JSON"
        >
            {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
        </Button>
        <pre className="text-[10px] md:text-xs font-mono bg-muted/50 p-3 rounded-md overflow-x-auto text-foreground/90 border">
            {JSON.stringify(data, null, 2)}
        </pre>
    </div>
  );
}
