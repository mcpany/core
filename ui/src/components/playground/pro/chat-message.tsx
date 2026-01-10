/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { User, Bot, Terminal, Sparkles, AlertCircle, Check, Copy } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Button } from "@/components/ui/button";
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { useState } from "react";

export type MessageType = "user" | "assistant" | "tool-call" | "tool-result" | "error";

export interface Message {
  id: string;
  type: MessageType;
  content?: string;
  toolName?: string;
  toolArgs?: Record<string, unknown>;
  toolResult?: unknown;
  timestamp: Date;
}

interface ChatMessageProps {
    message: Message;
}

export function ChatMessage({ message }: ChatMessageProps) {
    const [copied, setCopied] = useState(false);

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    if (message.type === "user") {
        return (
            <div className="flex justify-end gap-3 pl-10 group">
                <div className="flex flex-col items-end gap-1 max-w-full">
                     <div className="bg-primary text-primary-foreground rounded-2xl rounded-tr-sm px-4 py-2 shadow-sm max-w-full overflow-hidden">
                        <p className="whitespace-pre-wrap font-mono text-sm break-words">{message.content}</p>
                    </div>
                     <span className="text-[10px] text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" suppressHydrationWarning>
                        {message.timestamp.toLocaleTimeString()}
                    </span>
                </div>
                 <Avatar className="size-8 mt-1 border shadow-sm shrink-0">
                    <AvatarFallback className="bg-primary/10 text-primary"><User className="size-4" /></AvatarFallback>
                </Avatar>
            </div>
        );
    }

    if (message.type === "assistant") {
        return (
            <div className="flex justify-start gap-3 pr-10 group">
                <Avatar className="size-8 mt-1 border bg-muted shadow-sm shrink-0">
                    <AvatarFallback><Bot className="size-4 text-primary" /></AvatarFallback>
                </Avatar>
                <div className="flex flex-col items-start gap-1 max-w-full">
                    <div className="bg-muted/50 rounded-2xl rounded-tl-sm px-4 py-2 shadow-sm border max-w-full overflow-hidden">
                         <p className="whitespace-pre-wrap text-sm">{message.content}</p>
                    </div>
                    <span className="text-[10px] text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
                        {message.timestamp.toLocaleTimeString()}
                    </span>
                </div>
            </div>
        );
    }

    if (message.type === "tool-call") {
        return (
            <div className="flex justify-start gap-3 pl-11 w-full pr-4 md:pr-10 my-2">
                <Card className="w-full border border-primary/20 bg-primary/5 shadow-none overflow-hidden">
                    <CardHeader className="p-3 pb-2 flex flex-row items-center gap-2 space-y-0 border-b border-primary/10 bg-primary/10">
                         <div className="bg-primary/20 p-1.5 rounded-md">
                             <Terminal className="size-3.5 text-primary" />
                         </div>
                         <div className="flex flex-col">
                             <span className="text-[10px] text-primary/70 uppercase tracking-wider font-semibold">Tool Execution</span>
                             <span className="font-mono text-sm font-medium text-primary">{message.toolName}</span>
                         </div>
                    </CardHeader>
                    <CardContent className="p-0">
                         <div className="relative group/code">
                            <SyntaxHighlighter
                                language="json"
                                style={vscDarkPlus}
                                customStyle={{ margin: 0, padding: '1rem', fontSize: '12px', background: 'rgba(0,0,0,0.4)' }}
                                wrapLines={true}
                                wrapLongLines={true}
                            >
                                {JSON.stringify(message.toolArgs, null, 2)}
                            </SyntaxHighlighter>
                            <Button
                                size="icon"
                                variant="ghost"
                                className="absolute right-2 top-2 h-6 w-6 opacity-0 group-hover/code:opacity-100 transition-opacity bg-muted/20 hover:bg-muted/50 text-white"
                                onClick={() => copyToClipboard(JSON.stringify(message.toolArgs, null, 2))}
                            >
                                {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                            </Button>
                        </div>
                    </CardContent>
                </Card>
            </div>
        );
    }

    if (message.type === "tool-result") {
         return (
            <div className="flex justify-start gap-3 pl-11 w-full pr-4 md:pr-10 my-2">
                 <Collapsible className="w-full group rounded-lg border shadow-sm overflow-hidden bg-card" defaultOpen>
                    <div className="flex items-center justify-between bg-muted/30 px-3 py-2 text-xs border-b">
                        <div className="flex items-center gap-2 text-muted-foreground">
                            <div className="bg-green-500/10 p-1 rounded">
                                <Sparkles className="size-3.5 text-green-600 dark:text-green-400" />
                            </div>
                            <span className="font-medium text-green-700 dark:text-green-400">Result: {message.toolName}</span>
                        </div>
                        <CollapsibleTrigger asChild>
                            <Button variant="ghost" size="sm" className="h-6 px-2 text-[10px] text-muted-foreground hover:text-foreground">
                                <span className="group-data-[state=open]:hidden">Show Result</span>
                                <span className="group-data-[state=closed]:hidden">Hide Result</span>
                            </Button>
                        </CollapsibleTrigger>
                    </div>
                    <CollapsibleContent>
                        <div className="relative group/code max-h-[400px] overflow-auto">
                             <SyntaxHighlighter
                                language="json"
                                style={vscDarkPlus}
                                customStyle={{ margin: 0, padding: '1rem', fontSize: '12px', minHeight: '100%' }}
                                wrapLines={true}
                                wrapLongLines={true}
                            >
                                {JSON.stringify(message.toolResult, null, 2)}
                            </SyntaxHighlighter>
                             <Button
                                size="icon"
                                variant="ghost"
                                className="absolute right-2 top-2 h-6 w-6 opacity-0 group-hover/code:opacity-100 transition-opacity bg-muted/20 hover:bg-muted/50 text-white"
                                onClick={() => copyToClipboard(JSON.stringify(message.toolResult, null, 2))}
                            >
                                {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
                            </Button>
                        </div>
                    </CollapsibleContent>
                </Collapsible>
            </div>
        );
    }

    if (message.type === "error") {
        return (
             <div className="flex justify-start gap-3 pl-11 w-full pr-10 my-2">
                <div className="flex items-start gap-3 text-destructive bg-destructive/5 px-4 py-3 rounded-lg text-sm border border-destructive/20 shadow-sm w-full">
                    <AlertCircle className="size-5 mt-0.5 shrink-0" />
                    <div className="flex flex-col gap-1">
                        <span className="font-semibold text-xs uppercase tracking-wider">Execution Error</span>
                        <span className="whitespace-pre-wrap font-mono text-xs">{message.content}</span>
                    </div>
                </div>
            </div>
        );
    }

    return null;
}
