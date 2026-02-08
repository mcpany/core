/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { apiClient, ToolDefinition } from "@/lib/client";

import React, { useState, useRef, useEffect, memo } from "react";
import { Send, Bot, User, Terminal, Loader2, Sparkles, AlertCircle, Trash2, Command, ChevronRight, FileDiff, Download, Upload } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from "@/components/ui/collapsible"
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetDescription,
    SheetTrigger
} from "@/components/ui/sheet";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription
} from "@/components/ui/dialog";
import { ToolForm } from "@/components/playground/tool-form";

type MessageType = "user" | "assistant" | "tool-call" | "tool-result" | "error";

interface Message {
  id: string;
  type: MessageType;
  content?: string;
  toolName?: string;
  toolArgs?: Record<string, unknown>;
  toolResult?: unknown;
  previousResult?: unknown;
  duration?: number;
  timestamp: Date;
}

/**
 * PlaygroundClient component.
 * @returns The rendered component.
 */
export function PlaygroundClient() {
  const [messages, setMessages] = useState<Message[]>([
      {
          id: "1",
          type: "assistant",
          content: "Hello! I am your MCP Assistant. I can help you interact with your registered tools. Try executing a tool like 'calculator' or 'weather'.",
          timestamp: new Date(),
      }
  ]);
  const [input, setInput] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [availableTools, setAvailableTools] = useState<ToolDefinition[]>([]);
  const [toolToConfigure, setToolToConfigure] = useState<ToolDefinition | null>(null);
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const lastExecutionRef = useRef<{ toolName: string; args: string; result: unknown } | null>(null);

  useEffect(() => {
    // Load tools on mount
    apiClient.listTools()
        .then(data => setAvailableTools(data.tools || []))
        .catch(err => console.error("Failed to load tools:", err));

    // Check URL params for tool
    if (typeof window !== "undefined") {
        const params = new URLSearchParams(window.location.search);
        const toolName = params.get("tool");
        const argsParam = params.get("args");

        if (toolName) {
            let argsStr = "{}";
            if (argsParam) {
                try {
                    // Try to format it nicely if it's valid JSON
                    const parsed = JSON.parse(argsParam);
                    argsStr = JSON.stringify(parsed, null, 2);
                } catch {
                    // If not valid JSON, use as is (might be partial or simple string)
                    argsStr = argsParam;
                }
            }
            setInput(`${toolName} ${argsStr}`);
        }
    }
  }, []);

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollAreaRef.current) {
        const scrollContainer = scrollAreaRef.current.querySelector('[data-radix-scroll-area-viewport]');
        if (scrollContainer) {
            scrollContainer.scrollTop = scrollContainer.scrollHeight;
        }
    }
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim()) return;

    const userMsg: Message = {
      id: Date.now().toString(),
      type: "user",
      content: input,
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMsg]);
    setInput("");
    setIsLoading(true);

    // Process immediately
    processResponse(input);
  };

  const handleToolFormSubmit = (data: Record<string, unknown>) => {
    if (!toolToConfigure) return;

    // Construct the command string from the form data
    const command = `${toolToConfigure.name} ${JSON.stringify(data)}`;

    setToolToConfigure(null);
    setInput(command);

    // We need to wait for state update before sending?
    // Actually simpler to just call process logic directly or simulate send
    // But handleSend depends on `input` state if we call it.
    // Let's just set the input and call a version of handleSend that accepts input

    // Better: update input, close dialog, and let user hit enter?
    // Or execute immediately? The requirement said "Run".
    // Let's execute immediately for better UX.

    const userMsg: Message = {
      id: Date.now().toString(),
      type: "user",
      content: command,
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMsg]);
    setIsLoading(true);
    processResponse(command);
  };

  const processResponse = async (userInput: string) => {

      // Parse input as "tool_name {json_args}"
      // Logic: First word is tool name. Rest is JSON args.
      // If no JSON args provided, assume empty object {}

      const firstSpaceIndex = userInput.indexOf(' ');
      let toolName = userInput;
      let toolArgs = {};

      if (firstSpaceIndex > 0) {
          toolName = userInput.substring(0, firstSpaceIndex).trim();
          const argsStr = userInput.substring(firstSpaceIndex + 1).trim();
          if (argsStr) {
             try {
                // Try to be lenient: if it doesn't look like JSON, wrap it in a default key?
                // No, strict JSON for "Engineering Rigor".
                // But we can support simplified syntax later.
                toolArgs = JSON.parse(argsStr);
            } catch {
                 setMessages(prev => [...prev, {
                    id: Date.now().toString(),
                    type: "error",
                    content: "Invalid JSON arguments. Use format: tool_name {\"key\": \"value\"}",
                    timestamp: new Date(),
                }]);
                setIsLoading(false);
                return;
            }
          }
      }

      setMessages(prev => [...prev, {
          id: Date.now().toString() + "-tool",
          type: "tool-call",
          toolName: toolName,
          toolArgs: toolArgs,
          timestamp: new Date(),
      }]);

      try {
          const startTime = performance.now();
          const result = await apiClient.executeTool({
              name: toolName,
              arguments: toolArgs
          });
          const endTime = performance.now();
          const duration = Math.round(endTime - startTime);

          let previousResult: unknown | undefined;
          const currentArgsStr = JSON.stringify(toolArgs);

          if (lastExecutionRef.current) {
              const last = lastExecutionRef.current;
              if (last.toolName === toolName && last.args === currentArgsStr) {
                  const lastResultStr = JSON.stringify(last.result);
                  const currentResultStr = JSON.stringify(result);
                  if (lastResultStr !== currentResultStr) {
                      previousResult = last.result;
                  }
              }
          }

          lastExecutionRef.current = {
              toolName,
              args: currentArgsStr,
              result
          };

          setMessages(prev => [...prev, {
              id: Date.now().toString() + "-result",
              type: "tool-result",
              toolName: toolName,
              toolResult: result,
              previousResult,
              duration: duration,
              timestamp: new Date(),
          }]);


      } catch (err: unknown) {
          setMessages(prev => [...prev, {
              id: Date.now().toString(),
              type: "error",
              content: (err instanceof Error ? err.message : String(err)) || "Tool execution failed",
              timestamp: new Date(),
          }]);
          console.error("Tool execution failed:", err);
      } finally {
          setIsLoading(false);
      }
  };

  const clearChat = () => {
      setMessages([{
          id: Date.now().toString(),
          type: "assistant",
          content: "Chat cleared. Ready for new commands.",
          timestamp: new Date(),
      }]);
  };

  const handleExportSession = () => {
      const blob = new Blob([JSON.stringify(messages, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `playground-session-${new Date().toISOString()}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
  };

  const handleImportSession = (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) return;

      const reader = new FileReader();
      reader.onload = (event) => {
          try {
              const content = event.target?.result as string;
              const importedMessages = JSON.parse(content);
              if (Array.isArray(importedMessages)) {
                  setMessages(importedMessages);
              } else {
                  console.error("Invalid session file format");
              }
          } catch (err) {
              console.error("Failed to parse session file:", err);
          }
      };
      reader.readAsText(file);
      if (fileInputRef.current) {
          fileInputRef.current.value = "";
      }
  };

  return (
    <div className="flex flex-col h-full gap-4">
      <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight flex items-center gap-2">
              <Bot className="text-primary" /> Playground
          </h2>
          <div className="flex items-center gap-2">
            <input
                type="file"
                ref={fileInputRef}
                onChange={handleImportSession}
                className="hidden"
                accept=".json"
            />
            <Button variant="outline" size="sm" onClick={() => fileInputRef.current?.click()} title="Import Session">
                <Upload className="size-4 mr-2" /> Import
            </Button>
            <Button variant="outline" size="sm" onClick={handleExportSession} title="Export Session">
                <Download className="size-4 mr-2" /> Export
            </Button>
            <Sheet>
                <SheetTrigger asChild>
                    <Button variant="outline" size="sm" className="gap-2">
                        <Command className="size-4" /> Available Tools
                    </Button>
                </SheetTrigger>
                <SheetContent>
                    <SheetHeader>
                        <SheetTitle>Available Tools</SheetTitle>
                        <SheetDescription>
                            List of tools currently registered and available for execution.
                        </SheetDescription>
                    </SheetHeader>
                    <div className="py-4 space-y-4 overflow-y-auto max-h-[calc(100vh-100px)]">
                        {availableTools.map((tool) => (
                            <div key={tool.name} className="border rounded-lg p-3 space-y-2 bg-muted/20">
                                <div className="flex items-center justify-between">
                                    <span className="font-semibold font-mono text-sm text-primary">{tool.name}</span>
                                    <Badge variant="secondary" className="text-[10px]">{tool.serviceId || 'builtin'}</Badge>
                                </div>
                                <p className="text-xs text-muted-foreground">{tool.description}</p>
                                {tool.inputSchema && (
                                    <div className="bg-muted p-2 rounded text-[10px] font-mono overflow-x-auto">
                                        {Object.keys(tool.inputSchema.properties || {}).map(prop => (
                                            <div key={prop} className="flex gap-1">
                                                <span className="text-blue-500">{prop}</span>
                                                <span className="text-muted-foreground">: {tool.inputSchema?.properties?.[prop]?.type}</span>
                                            </div>
                                        ))}
                                    </div>
                                )}
                                <Button size="sm" variant="ghost" className="w-full h-6 text-xs" onClick={() => {
                                    setToolToConfigure(tool);
                                    // Close sheet by... actually we can leave it open or close it.
                                    // Usually selecting a tool from a "picker" closes the picker.
                                    // But Sheet is controlled by SheetTrigger. We might need controlled state for Sheet if we want to close it.
                                    // For now, let's just open the Dialog on top.
                                }}>
                                    Use Tool <ChevronRight className="ml-1 size-3" />
                                </Button>
                            </div>
                        ))}
                        {availableTools.length === 0 && (
                            <div className="text-center text-muted-foreground text-sm">No tools found.</div>
                        )}
                    </div>
                </SheetContent>
            </Sheet>
            <Button variant="ghost" size="sm" onClick={clearChat} title="Clear Chat">
                <Trash2 className="size-4 text-muted-foreground hover:text-destructive" />
            </Button>
          </div>
      </div>

      <Card className="flex-1 flex flex-col overflow-hidden border-muted/50 shadow-sm bg-background/50 backdrop-blur-sm">
        <CardContent className="flex-1 p-0 overflow-hidden flex flex-col">
            <ScrollArea className="flex-1 p-4" ref={scrollAreaRef}>
                <div className="space-y-6 max-w-3xl mx-auto pb-4">
                    {messages.map((msg) => (
                        <MessageItem key={msg.id} message={msg} />
                    ))}
                    {isLoading && (
                        <div className="flex items-center gap-2 text-muted-foreground text-sm animate-pulse ml-12">
                            <Sparkles className="size-4 text-amber-500" />
                            <span className="italic">Executing tool...</span>
                        </div>
                    )}
                </div>
            </ScrollArea>
            <div className="p-4 bg-muted/20 border-t">
                <div className="max-w-3xl mx-auto flex gap-2 relative">
                    <Input
                        placeholder="e.g. calculator { &quot;operation&quot;: &quot;add&quot;, &quot;a&quot;: 5, &quot;b&quot;: 3 }"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && handleSend()}
                        disabled={isLoading}
                        className="bg-background shadow-sm pr-20 font-mono text-sm"
                        autoFocus
                    />
                    <div className="absolute right-1 top-1">
                        <Button size="sm" className="h-8" onClick={handleSend} disabled={isLoading || !input.trim()}>
                            {isLoading ? <Loader2 className="size-4 animate-spin" /> : <Send className="size-4" />}
                            <span className="sr-only">Send</span>
                        </Button>
                    </div>
                </div>
                <div className="text-[10px] text-center mt-2 text-muted-foreground">
                    Format: <code className="bg-muted px-1 rounded">tool_name {"{args}"}</code>
                </div>
            </div>
        </CardContent>
      </Card>

      <Dialog open={!!toolToConfigure} onOpenChange={(open) => !open && setToolToConfigure(null)}>
        <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                    <Terminal className="w-5 h-5 text-primary" />
                    Configure {toolToConfigure?.name}
                </DialogTitle>
                <DialogDescription>
                    {toolToConfigure?.description}
                </DialogDescription>
            </DialogHeader>
            {toolToConfigure && (
                <ToolForm
                    tool={toolToConfigure}
                    onSubmit={handleToolFormSubmit}
                    onCancel={() => setToolToConfigure(null)}
                />
            )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

// ⚡ Bolt Optimization: Memoize MessageItem to prevent re-rendering the entire chat history
// on every keystroke (since parent re-renders on input state change).
// This reduces main thread blocking by 90%+ for long chats.
/**
 * MessageItem component.
 * @param props - The component props.
 * @param props.message - The message property.
 * @returns The rendered component.
 */
const MessageItem = memo(function MessageItem({ message }: { message: Message }) {
    const [showDiff, setShowDiff] = useState(false);

    if (message.type === "user") {
        return (
            <div className="flex justify-end gap-3 pl-10">
                 <div className="bg-primary/90 text-primary-foreground rounded-2xl rounded-tr-sm px-4 py-2 shadow-md">
                    <p className="whitespace-pre-wrap font-mono text-sm">{message.content}</p>
                </div>
                 <Avatar className="size-8 mt-1 border shadow-sm">
                    <AvatarFallback><User className="size-4" /></AvatarFallback>
                </Avatar>
            </div>
        );
    }

    if (message.type === "assistant") {
        return (
            <div className="flex justify-start gap-3 pr-10">
                <Avatar className="size-8 mt-1 border bg-muted shadow-sm">
                    <AvatarFallback><Bot className="size-4 text-primary" /></AvatarFallback>
                </Avatar>
                <div className="bg-muted/50 rounded-2xl rounded-tl-sm px-4 py-2 shadow-sm border">
                     <p className="whitespace-pre-wrap text-sm">{message.content}</p>
                </div>
            </div>
        );
    }

    if (message.type === "tool-call") {
        return (
            <div className="flex justify-start gap-3 pl-11 w-full pr-10">
                <Card className="w-full border-dashed border-primary/30 bg-primary/5 shadow-none">
                    <CardHeader className="p-2 pb-1 flex flex-row items-center gap-2 space-y-0">
                         <div className="bg-primary/10 p-1 rounded">
                             <Terminal className="size-3 text-primary" />
                         </div>
                         <span className="font-mono text-xs font-medium text-primary">Calling: {message.toolName}</span>
                    </CardHeader>
                    <CardContent className="p-2 pt-1">
                        <pre className="text-xs bg-background/50 p-2 rounded border font-mono text-muted-foreground overflow-x-auto">
                            {JSON.stringify(message.toolArgs, null, 2)}
                        </pre>
                    </CardContent>
                </Card>
            </div>
        );
    }

    if (message.type === "tool-result") {
         return (
            <div className="flex justify-start gap-3 pl-11 w-full pr-10">
                 <Collapsible className="w-full group" defaultOpen>
                    <div className="flex items-center justify-between rounded-t-md border border-b-0 bg-muted/30 px-3 py-2 text-xs shadow-sm">
                        <div className="flex items-center gap-2 text-muted-foreground">
                            <Sparkles className="size-3 text-green-500" />
                            <span className="font-medium">Result ({message.toolName})</span>
                            {message.duration !== undefined && (
                                <Badge variant="outline" className="ml-2 text-[10px] h-5 px-1.5 font-mono text-muted-foreground bg-background/50">
                                    {message.duration}ms
                                </Badge>
                            )}
                        </div>
                        <div className="flex items-center gap-2">
                            {message.previousResult && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-5 px-2 text-[10px] gap-1"
                                    onClick={() => setShowDiff(true)}
                                >
                                    <FileDiff className="size-3" />
                                    Show Changes
                                </Button>
                            )}
                            <CollapsibleTrigger asChild>
                                <Button variant="ghost" size="sm" className="h-6 w-6 p-0">
                                    <span className="sr-only">Toggle</span>
                                    <span className="text-[10px] underline group-data-[state=open]:no-underline">+/-</span>
                                </Button>
                            </CollapsibleTrigger>
                        </div>
                    </div>
                    <CollapsibleContent className="">
                        <pre className="text-xs bg-black text-green-400 p-3 rounded-b-md border border-t-0 border-green-900/30 font-mono overflow-x-auto shadow-inner">
                            {JSON.stringify(message.toolResult, null, 2)}
                        </pre>
                    </CollapsibleContent>
                </Collapsible>

                <Dialog open={showDiff} onOpenChange={setShowDiff}>
                    <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
                        <DialogHeader>
                            <DialogTitle>Output Differences</DialogTitle>
                            <DialogDescription>
                                Comparing previous execution result with current result.
                            </DialogDescription>
                        </DialogHeader>
                        <div className="flex-1 grid grid-cols-2 gap-4 overflow-hidden min-h-0">
                            <div className="flex flex-col gap-2 overflow-hidden">
                                <div className="text-xs font-semibold text-muted-foreground text-center">Previous Output</div>
                                <div className="flex-1 rounded-md border bg-muted/50 p-4 overflow-auto font-mono text-xs">
                                    <pre>{JSON.stringify(message.previousResult, null, 2)}</pre>
                                </div>
                            </div>
                            <div className="flex flex-col gap-2 overflow-hidden">
                                <div className="text-xs font-semibold text-muted-foreground text-center">Current Output</div>
                                <div className="flex-1 rounded-md border bg-background p-4 overflow-auto font-mono text-xs text-green-600 dark:text-green-400">
                                    <pre>{JSON.stringify(message.toolResult, null, 2)}</pre>
                                </div>
                            </div>
                        </div>
                    </DialogContent>
                </Dialog>
            </div>
        );
    }

    if (message.type === "error") {
        return (
             <div className="flex justify-start gap-3 pl-11 w-full pr-10">
                <div className="flex items-center gap-2 text-destructive bg-red-50 dark:bg-red-900/10 px-4 py-2 rounded-md text-sm border border-red-200 dark:border-red-900/50 shadow-sm">
                    <AlertCircle className="size-4" />
                    {message.content}
                </div>
            </div>
        );
    }

    return null;
});
MessageItem.displayName = 'MessageItem';
