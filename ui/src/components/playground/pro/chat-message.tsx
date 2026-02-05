/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { User, Bot, Terminal, Sparkles, AlertCircle, Check, Copy, RotateCcw, Lightbulb, GitCompare } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { useState, useEffect } from "react";
import { SmartResultRenderer } from "./smart-result-renderer";
import { estimateTokens, formatTokenCount } from "@/lib/tokens";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useTheme } from "next-themes";
import { defineDraculaTheme } from "@/lib/monaco-theme";
import dynamic from "next/dynamic";

// ⚡ BOLT: Lazy load heavy dependencies to improve initial bundle size and TTI.
// Randomized Selection from Top 5 High-Impact Targets
const SyntaxHighlighter = dynamic(
    () => import("react-syntax-highlighter").then((mod) => mod.Prism),
    {
        ssr: false,
        loading: () => <div className="p-4 bg-[rgba(0,0,0,0.4)] h-12 animate-pulse rounded" />,
    }
);

const DiffEditor = dynamic(
    () => import("@monaco-editor/react").then((mod) => mod.DiffEditor),
    {
        ssr: false,
        loading: () => <div className="h-full w-full bg-[#1e1e1e] animate-pulse rounded-md" />,
    }
);

/**
 * Defines the possible types of messages in the chat interface.
 */
export type MessageType = "user" | "assistant" | "tool-call" | "tool-result" | "error";

/**
 * Represents a single message object in the chat history.
 */
export interface Message {
  /** Unique ID for the message. */
  id: string;
  /** The type of message (user, assistant, tool, error). */
  type: MessageType;
  /** The textual content of the message. */
  content?: string;
  /** The name of the tool being called (if applicable). */
  toolName?: string;
  /** Arguments passed to the tool (if applicable). */
  toolArgs?: Record<string, unknown>;
  /** The result returned by the tool (if applicable). */
  toolResult?: unknown;
  /** The previous result for diffing purposes. */
  previousResult?: unknown;
  /** The timestamp when the message was created. */
  timestamp: Date;
}

interface ChatMessageProps {
    message: Message;
    onReplay?: (toolName: string, args: Record<string, unknown>) => void;
    onRetry?: (toolName: string, args: Record<string, unknown>) => void;
}

function analyzeError(error: string): string | null {
    const e = error.toLowerCase();
    if (e.includes("timeout") || e.includes("timed out") || e.includes("deadline")) {
        return "The tool took too long to respond. Try increasing the timeout or checking if the service is overloaded.";
    }
    if (e.includes("json") || e.includes("parse") || e.includes("syntax")) {
        return "It looks like there's a syntax error in your arguments. Please check the JSON format.";
    }
    if (e.includes("connection") || e.includes("network") || e.includes("refused")) {
        return "Could not connect to the upstream service. Please check if the service is running and reachable.";
    }
    if (e.includes("unauthorized") || e.includes("authentication") || e.includes("401") || e.includes("403")) {
        return "Authentication failed. Please check your API keys or credentials in the service configuration.";
    }
    if (e.includes("not found") || e.includes("404")) {
        return "The requested tool or resource was not found. Please verify the tool name and availability.";
    }
    if (e.includes("argument") || e.includes("required") || e.includes("validation") || e.includes("missing")) {
        return "Some required arguments might be missing or invalid. Please check the tool schema.";
    }
    return null;
}

/**
 * Renders a single chat message based on its type.
 *
 * Handles various message types including:
 * - User messages (right aligned)
 * - Assistant messages (left aligned)
 * - Tool calls (with syntax highlighting and copy support)
 * - Tool results (with diff viewer and collapsible content)
 * - Errors (with analysis suggestions)
 *
 * @param props - The component props.
 * @param props.message - The message object to display.
 * @param props.onReplay - Callback when a tool call is replayed/loaded into console.
 * @param props.onRetry - Callback when a failed tool call is retried.
 * @returns The rendered chat message component.
 */
export function ChatMessage({ message, onReplay, onRetry }: ChatMessageProps) {
    const [copied, setCopied] = useState(false);
    const [showDiff, setShowDiff] = useState(false);
    const { theme } = useTheme();

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
                     <span className="text-[10px] text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
                        <HydrationSafeTime date={message.timestamp} />
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
                         <HydrationSafeTime date={message.timestamp} />
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
                         <div className="flex-1" />
                         <span className="text-[10px] text-muted-foreground mr-2" title="Estimated token usage for arguments">
                            {formatTokenCount(estimateTokens(JSON.stringify(message.toolArgs || {})))} tokens
                         </span>
                         {onReplay && message.toolName && (
                             <Tooltip>
                                 <TooltipTrigger asChild>
                                      <Button
                                         variant="ghost"
                                         size="icon"
                                         className="h-6 w-6 text-muted-foreground hover:text-foreground"
                                         onClick={() => onReplay(message.toolName!, message.toolArgs || {})}
                                         aria-label="Load into console"
                                      >
                                          <RotateCcw className="h-3.5 w-3.5" />
                                      </Button>
                                 </TooltipTrigger>
                                 <TooltipContent>
                                     <p>Load into console</p>
                                 </TooltipContent>
                             </Tooltip>
                         )}
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
        const hasDiff = message.previousResult !== undefined &&
                        JSON.stringify(message.previousResult) !== JSON.stringify(message.toolResult);

         return (
            <>
            <div className="flex justify-start gap-3 pl-11 w-full pr-4 md:pr-10 my-2">
                 <Collapsible className="w-full group rounded-lg border shadow-sm overflow-hidden bg-card" defaultOpen>
                    <div className="flex items-center justify-between bg-muted/30 px-3 py-2 text-xs border-b">
                        <div className="flex items-center gap-2 text-muted-foreground">
                            <div className="bg-green-500/10 p-1 rounded">
                                <Sparkles className="size-3.5 text-green-600 dark:text-green-400" />
                            </div>
                            <span className="font-medium text-green-700 dark:text-green-400">Result: {message.toolName}</span>
                            <span className="ml-2 text-[10px] opacity-60" title="Estimated token usage for result">
                                ({formatTokenCount(estimateTokens(JSON.stringify(message.toolResult)))} tokens)
                            </span>
                        </div>
                        <div className="flex items-center gap-2">
                            {hasDiff && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-6 px-2 text-[10px] gap-1 border-dashed"
                                    onClick={() => setShowDiff(true)}
                                >
                                    <GitCompare className="size-3" />
                                    Show Changes
                                </Button>
                            )}
                            <CollapsibleTrigger asChild>
                                <Button variant="ghost" size="sm" className="h-6 px-2 text-[10px] text-muted-foreground hover:text-foreground">
                                    <span className="group-data-[state=open]:hidden">Show Result</span>
                                    <span className="group-data-[state=closed]:hidden">Hide Result</span>
                                </Button>
                            </CollapsibleTrigger>
                        </div>
                    </div>
                    <CollapsibleContent>
                        <SmartResultRenderer result={message.toolResult} />
                    </CollapsibleContent>
                </Collapsible>
            </div>

            <Dialog open={showDiff} onOpenChange={setShowDiff}>
                <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <GitCompare className="size-5" />
                            Output Difference
                        </DialogTitle>
                    </DialogHeader>
                    <div className="flex-1 border rounded-md overflow-hidden bg-[#1e1e1e]">
                        <DiffEditor
                            original={JSON.stringify(message.previousResult, null, 2)}
                            modified={JSON.stringify(message.toolResult, null, 2)}
                            language="json"
                            theme={theme === "dark" ? "dracula" : "light"}
                            onMount={(editor, monaco) => {
                                if (theme === "dark") {
                                    defineDraculaTheme(monaco);
                                    monaco.editor.setTheme("dracula");
                                }
                            }}
                            options={{
                                readOnly: true,
                                minimap: { enabled: false },
                                scrollBeyondLastLine: false,
                                fontSize: 12,
                                diffCodeLens: true,
                                renderSideBySide: true,
                            }}
                        />
                    </div>
                </DialogContent>
            </Dialog>
            </>
        );
    }

    if (message.type === "error") {
        const suggestion = message.content ? analyzeError(message.content) : null;

        return (
             <div className="flex justify-start gap-3 pl-11 w-full pr-10 my-2">
                <div className="flex flex-col gap-2 w-full">
                    <div className="flex items-start gap-3 text-destructive bg-destructive/5 px-4 py-3 rounded-lg text-sm border border-destructive/20 shadow-sm w-full relative group">
                        <AlertCircle className="size-5 mt-0.5 shrink-0" />
                        <div className="flex flex-col gap-1 flex-1">
                            <span className="font-semibold text-xs uppercase tracking-wider">Execution Error</span>
                            <span className="whitespace-pre-wrap font-mono text-xs break-all">{message.content}</span>
                        </div>
                        {onRetry && message.toolName && (
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 -mt-1 -mr-2 text-destructive/70 hover:text-destructive hover:bg-destructive/10"
                                        onClick={() => onRetry(message.toolName!, message.toolArgs || {})}
                                        aria-label="Retry command"
                                    >
                                        <RotateCcw className="h-3.5 w-3.5" />
                                    </Button>
                                </TooltipTrigger>
                                <TooltipContent side="left">
                                    <p>Retry this command</p>
                                </TooltipContent>
                            </Tooltip>
                        )}
                    </div>

                    {suggestion && (
                        <div className="flex items-start gap-3 text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-900/10 px-4 py-3 rounded-lg text-sm border border-amber-200 dark:border-amber-800/30 shadow-sm w-full">
                            <Lightbulb className="size-5 mt-0.5 shrink-0" />
                            <div className="flex flex-col gap-1">
                                <span className="font-semibold text-xs uppercase tracking-wider">Suggestion</span>
                                <span className="text-xs">{suggestion}</span>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        );
    }

    return null;
}

/**
 * HydrationSafeTime component.
 * @param props - The component props.
 * @param props.date - The date property.
 * @returns The rendered component.
 */
function HydrationSafeTime({ date }: { date: Date }) {
    const [mounted, setMounted] = useState(false);
    useEffect(() => setMounted(true), []);
    if (!mounted) return null;
    return <>{date.toLocaleTimeString()}</>;
}
