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
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

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

  const jsonString = JSON.stringify(data, null, 2);

  return (
    <div className={cn("relative group rounded-md overflow-hidden border bg-[#1e1e1e]", className)}>
        <Button
            variant="ghost"
            size="icon"
            className="absolute right-2 top-2 h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity bg-white/10 hover:bg-white/20 text-white z-10"
            onClick={handleCopy}
            title="Copy JSON"
        >
            {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
        </Button>
        <SyntaxHighlighter
            language="json"
            style={vscDarkPlus}
            customStyle={{
                margin: 0,
                padding: '1rem',
                fontSize: '0.75rem', // text-xs
                lineHeight: '1.5',
                background: 'transparent', // Let parent handle bg
            }}
            wrapLines={true}
            wrapLongLines={true}
        >
            {jsonString}
        </SyntaxHighlighter>
    </div>
  );
}
