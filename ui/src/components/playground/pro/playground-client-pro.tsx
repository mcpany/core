/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { apiClient, ToolDefinition, PromptDefinition } from "@/lib/client";

import { useState, useRef, useEffect, useMemo } from "react";
import { Send, Loader2, Sparkles, Terminal, PanelLeftClose, PanelLeftOpen, Zap, MessageSquare } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Download, Share2, Copy, Check, Info, Upload } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { estimateTokens, estimateMessageTokens } from "@/lib/tokens";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription
} from "@/components/ui/dialog";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable"

import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";

import { ToolForm } from "@/components/playground/tool-form";
import { ToolSidebar } from "./tool-sidebar";
import { ChatMessage, Message } from "./chat-message";
import { useIsMobile } from "@/hooks/use-mobile";
import { useLocalStorage } from "@/hooks/use-local-storage";

import { useSearchParams } from "next/navigation";

/**
 * PlaygroundClientPro component.
 * @returns The rendered component.
 */
export function PlaygroundClientPro() {
  const [messages, setMessages, isInitialized] = useLocalStorage<Message[]>("playground-messages", []);
  const [input, setInput] = useState("");
  const searchParams = useSearchParams();

  // Initialize with welcome message if empty and only after local storage is loaded
  useEffect(() => {
    if (isInitialized) {
        // We only add welcome message if the array is empty AND
        // we check if the key was missing from localStorage (implies first visit).
        // However, useLocalStorage handles default value if key is missing.
        // If the user explicitly clears the chat, we set it to empty array.
        // So `messages` being empty array means either:
        // 1. First visit (default value used)
        // 2. User cleared chat

        // To strictly follow "persistence", if the user cleared it, it should stay cleared.
        // But for UX, if I open the page and it's empty, a welcome message is nice.
        // Let's rely on checking if localStorage has the key to distinguish first visit.
        const hasKey = typeof window !== "undefined" && window.localStorage.getItem("playground-messages") !== null;

        if (!hasKey && messages.length === 0) {
            setMessages([
                {
                    id: "1",
                    type: "assistant",
                    content: "Hello! I am your MCP Assistant. Select a tool from the sidebar to configure and execute it, or type a command directly.",
                    timestamp: new Date(),
                }
            ]);
        }
    }
  }, [isInitialized]); // Only run when initialization status changes

  // Revive dates from stored messages (JSON strings)
  const displayMessages = useMemo(() => {
      return messages.map(m => ({
          ...m,
          timestamp: new Date(m.timestamp)
      }));
  }, [messages]);

  useEffect(() => {
    const tool = searchParams.get('tool');
    const args = searchParams.get('args');
    if (tool) {
        let command = tool;
        if (args) {
             command += ` ${args}`;
        }
        setInput(command);
    }
  }, [searchParams]);

  const [isLoading, setIsLoading] = useState(false);
  const [availableTools, setAvailableTools] = useState<ToolDefinition[]>([]);
  const [availablePrompts, setAvailablePrompts] = useState<PromptDefinition[]>([]);
  const [toolToConfigure, setToolToConfigure] = useState<ToolDefinition | null>(null);
  const [promptToConfigure, setPromptToConfigure] = useState<PromptDefinition | null>(null);
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const isMobile = useIsMobile();
  const [isDryRun, setIsDryRun] = useState(false);
  const [copied, setCopied] = useState(false);
  const { toast } = useToast();

  const currentTokens = useMemo(() => estimateTokens(input), [input]);
  const historyTokens = useMemo(() => estimateMessageTokens(displayMessages), [displayMessages]);

  // Autocomplete state
  const [filteredSuggestions, setFilteredSuggestions] = useState<ToolDefinition[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);

  useEffect(() => {
    apiClient.listTools()
        .then(data => setAvailableTools(data.tools || []))
        .catch(err => console.error("Failed to load tools:", err));

    apiClient.listPrompts()
        .then(data => setAvailablePrompts(data.prompts || []))
        .catch(err => console.error("Failed to load prompts:", err));
  }, []);

  useEffect(() => {
      if (isMobile) setSidebarOpen(false);
  }, [isMobile]);

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollAreaRef.current) {
        const scrollContainer = scrollAreaRef.current.querySelector('[data-radix-scroll-area-viewport]');
        if (scrollContainer) {
            scrollContainer.scrollTop = scrollContainer.scrollHeight;
        }
    }
  }, [displayMessages]);

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
    setShowSuggestions(false);

    processResponse(input);
  };

  const handleToolFormSubmit = (data: Record<string, unknown>) => {
    if (!toolToConfigure) return;
    const command = `${toolToConfigure.name} ${JSON.stringify(data)}`;
    setToolToConfigure(null);
    setInput(command);
  };

  const handlePromptFormSubmit = async (data: Record<string, unknown>) => {
    if (!promptToConfigure) return;
    const name = promptToConfigure.name;
    setPromptToConfigure(null);
    setIsLoading(true);

    // Add a system note that we are running a prompt
    setMessages(prev => [...prev, {
        id: Date.now().toString() + "-prompt-req",
        type: "user",
        content: `Running prompt: ${name}`,
        timestamp: new Date(),
    }]);

    try {
        const result = await apiClient.executePrompt(name, data as Record<string, string>);

        const newMessages: Message[] = (result.messages || []).map((msg: any, idx: number) => ({
             id: Date.now().toString() + `-prompt-${idx}`,
             type: msg.role === "user" ? "user" : "assistant",
             content: typeof msg.content === 'string' ? msg.content : msg.content?.text || JSON.stringify(msg.content),
             timestamp: new Date()
        }));

        setMessages(prev => [...prev, ...newMessages]);
    } catch (e: any) {
        setMessages(prev => [...prev, {
            id: Date.now().toString(),
            type: "error",
            content: e.message || "Failed to execute prompt",
            timestamp: new Date()
        }]);
    } finally {
        setIsLoading(false);
    }
  };

  const handleInputChange = (value: string) => {
      setInput(value);
      if (value.trim()) {
          const suggestions = availableTools.filter(t => t.name.toLowerCase().includes(value.toLowerCase()));
          setFilteredSuggestions(suggestions);
          setShowSuggestions(suggestions.length > 0);
      } else {
          setShowSuggestions(false);
      }
  };

  const selectSuggestion = (tool: ToolDefinition) => {
      setToolToConfigure(tool);
      setShowSuggestions(false);
  };

  const handleReplay = (toolName: string, args: Record<string, unknown>) => {
      const command = `${toolName} ${JSON.stringify(args)}`;
      setInput(command);
      inputRef.current?.focus();
  };

  const processResponse = async (userInput: string) => {
      const firstSpaceIndex = userInput.indexOf(' ');
      let toolName = userInput;
      let toolArgs = {};

      if (firstSpaceIndex > 0) {
          toolName = userInput.substring(0, firstSpaceIndex).trim();
          const argsStr = userInput.substring(firstSpaceIndex + 1).trim();
          if (argsStr) {
             try {
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
          const result = await apiClient.executeTool({
              name: toolName,
              arguments: toolArgs
          }, isDryRun);

          // Find previous execution for diffing
          let previousResult: unknown | undefined;
          const reversedMessages = [...messages].reverse();
          const previousCall = reversedMessages.find(m =>
              m.type === "tool-call" &&
              m.toolName === toolName &&
              JSON.stringify(m.toolArgs) === JSON.stringify(toolArgs)
          );

          if (previousCall) {
              const callIndex = messages.findIndex(m => m.id === previousCall.id);
              if (callIndex !== -1 && callIndex + 1 < messages.length) {
                  const resultMsg = messages[callIndex + 1];
                  if (resultMsg.type === "tool-result") {
                      previousResult = resultMsg.toolResult;
                  }
              }
          }

          setMessages(prev => [...prev, {
              id: Date.now().toString() + "-result",
              type: "tool-result",
              toolName: toolName,
              toolResult: result,
              previousResult,
              timestamp: new Date(),
          }]);

      } catch (err: unknown) {
          setMessages(prev => [...prev, {
              id: Date.now().toString(),
              type: "error",
              content: (err instanceof Error ? err.message : String(err)) || "Tool execution failed",
              toolName: toolName,
              toolArgs: toolArgs,
              timestamp: new Date(),
          }]);
      } finally {
          setIsLoading(false);
      }
  };

  const handleExportHistory = () => {
    const data = JSON.stringify(messages, null, 2);
    const blob = new Blob([data], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `playground-history-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    toast({
        title: "History Exported",
        description: "Your playground session has been saved to a JSON file."
    });
  };

  const handleImportClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const content = e.target?.result as string;
        const importedMessages = JSON.parse(content);

        // Basic validation
        if (!Array.isArray(importedMessages)) {
            throw new Error("Invalid format: Root must be an array");
        }

        setMessages(importedMessages);
        toast({
            title: "History Imported",
            description: `Successfully loaded ${importedMessages.length} messages.`
        });
      } catch (err) {
        toast({
            title: "Import Failed",
            description: "Failed to parse the file. Ensure it is a valid JSON export.",
            variant: "destructive"
        });
        console.error("Import error:", err);
      }
    };
    reader.readAsText(file);
    // Reset input so same file can be selected again if needed
    event.target.value = '';
  };

  const handleShareUrl = () => {
      const url = new URL(window.location.href);
      // If the input starts with a tool name, we can try to parse it
      const parts = input.trim().split(/\s+(.*)/);
      if (parts[0] && availableTools.some(t => t.name === parts[0])) {
          url.searchParams.set("tool", parts[0]);
          if (parts[1]) {
              url.searchParams.set("args", parts[1]);
          }
      }

      navigator.clipboard.writeText(url.toString());
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);

      toast({
          title: "URL Copied",
          description: "A sharable link with your current tool configuration has been copied to clipboard."
      });
  };

  return (
    <div className="flex flex-col h-full bg-background">
      <ResizablePanelGroup direction="horizontal" className="h-full items-stretch">
         <ResizablePanel
            defaultSize={25}
            minSize={20}
            maxSize={40}
            collapsible={true}
            collapsedSize={0}
            className={!sidebarOpen ? "hidden" : ""}
            onCollapse={() => setSidebarOpen(false)}
            onExpand={() => setSidebarOpen(true)}
         >
             <ToolSidebar
                tools={availableTools}
                prompts={availablePrompts}
                onSelectTool={setToolToConfigure}
                onSelectPrompt={setPromptToConfigure}
             />
         </ResizablePanel>

         <ResizableHandle withHandle={!isMobile} className={!sidebarOpen ? "hidden" : ""} />

         <ResizablePanel defaultSize={75}>
            <div className="flex flex-col h-full relative bg-muted/5">
                {/* Header */}
                <div className="h-14 border-b flex items-center justify-between px-4 bg-background/80 backdrop-blur-sm sticky top-0 z-10">
                     <div className="flex items-center gap-2">
                        {!sidebarOpen && (
                             <Button variant="ghost" size="icon" onClick={() => setSidebarOpen(true)} className="h-8 w-8">
                                 <PanelLeftOpen className="h-4 w-4" />
                             </Button>
                        )}
                        {sidebarOpen && isMobile && (
                            <Button variant="ghost" size="icon" onClick={() => setSidebarOpen(false)} className="h-8 w-8">
                                <PanelLeftClose className="h-4 w-4" />
                            </Button>
                        )}
                        <h2 className="font-semibold text-sm flex items-center gap-2">
                            <Terminal className="h-4 w-4 text-primary" />
                            Console
                        </h2>
                     </div>
                     <div className="flex items-center gap-2">
                          <Button
                               variant="outline"
                               size="sm"
                               className="h-7 text-xs flex items-center gap-1"
                               onClick={handleShareUrl}
                          >
                              {copied ? <Check className="h-3 w-3" /> : <Share2 className="h-3 w-3" />}
                              Share
                          </Button>
                          <Button
                               variant="outline"
                               size="sm"
                               className="h-7 text-xs flex items-center gap-1"
                               onClick={handleExportHistory}
                               disabled={displayMessages.length === 0}
                          >
                              <Download className="h-3 w-3" />
                              Export
                          </Button>
                          <Button
                               variant="outline"
                               size="sm"
                               className="h-7 text-xs flex items-center gap-1"
                               onClick={handleImportClick}
                          >
                              <Upload className="h-3 w-3" />
                              Import
                          </Button>
                          <input
                                type="file"
                                ref={fileInputRef}
                                className="hidden"
                                accept=".json"
                                onChange={handleFileChange}
                          />
                          <Button
                            variant="outline"
                            size="sm"
                            className="h-7 text-xs"
                            onClick={() => setMessages([])}
                            disabled={displayMessages.length === 0}
                          >
                              Clear
                          </Button>
                     </div>
                </div>

                {/* Chat Area */}
                <div className="flex-1 overflow-hidden relative">
                    <ScrollArea className="h-full p-4" ref={scrollAreaRef}>
                        <div className="max-w-4xl mx-auto pb-10 space-y-4">
                            {displayMessages.map((msg) => (
                                <ChatMessage key={msg.id} message={msg} onReplay={handleReplay} onRetry={handleReplay} />
                            ))}
                            {isLoading && (
                                <div className="flex items-center gap-2 text-muted-foreground text-xs animate-pulse pl-12">
                                    <Sparkles className="size-3 text-primary" />
                                    <span className="italic">Processing execution...</span>
                                </div>
                            )}
                            <div className="h-4" /> {/* Spacer */}
                        </div>
                    </ScrollArea>
                </div>

                {/* Input Area */}
                <div className="p-4 bg-background border-t">
                    <div className="max-w-4xl mx-auto flex gap-3 relative">
                         <div className="flex-1 relative">
                            <Input
                                ref={inputRef}
                                placeholder="Enter command or select a tool..."
                                value={input}
                                onChange={(e) => handleInputChange(e.target.value)}
                                onKeyDown={(e) => e.key === "Enter" && handleSend()}
                                disabled={isLoading}
                                className="pr-12 font-mono text-sm bg-muted/20 focus-visible:bg-background transition-colors h-11"
                                autoFocus
                            />
                            {showSuggestions && (
                                <div className="absolute bottom-full left-0 w-full bg-popover border rounded-md shadow-md mb-2 overflow-hidden z-20">
                                    <div className="p-1">
                                        {filteredSuggestions.map(tool => (
                                            <div
                                                key={tool.name}
                                                className="px-2 py-1.5 text-sm cursor-pointer hover:bg-accent hover:text-accent-foreground rounded-sm flex items-center justify-between"
                                                onClick={() => selectSuggestion(tool)}
                                            >
                                                <span className="font-medium">{tool.name}</span>
                                                <span className="text-xs text-muted-foreground">{tool.serviceId || 'core'}</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}
                             <div className="absolute right-1 top-1.5">
                                <Button
                                    size="sm"
                                    className="h-8 w-8 p-0 rounded-md"
                                    onClick={handleSend}
                                    disabled={isLoading || !input.trim()}
                                    aria-label="Send"
                                >
                                    {isLoading ? <Loader2 className="size-4 animate-spin" /> : <Send className="size-4" />}
                                </Button>
                            </div>
                         </div>
                    </div>
                     <div className="max-w-4xl mx-auto mt-2 flex justify-between items-center text-[10px] text-muted-foreground px-1">
                        <div className="flex items-center gap-4">
                            <span>Format: <code className="bg-muted px-1 rounded text-primary">tool_name {"{json_args}"}</code></span>
                            <div className="flex items-center gap-1.5 border-l pl-4">
                                <Switch id="console-dry-run" checked={isDryRun} onCheckedChange={setIsDryRun} className="scale-75 origin-left" />
                                <Label htmlFor="console-dry-run" className="cursor-pointer text-[10px]">Dry Run</Label>
                            </div>
                            <div className="flex items-center gap-1.5 border-l pl-4">
                                <Info className="h-3 w-3 text-muted-foreground" />
                                <span title="Approximate tokens based on character count and words">
                                    ~{currentTokens} tokens
                                </span>
                                <span className="text-[9px] opacity-60 ml-1">
                                   (Session: ~{historyTokens})
                                </span>
                            </div>
                        </div>
                        <span className="hidden sm:inline">Press Enter to execute</span>
                    </div>
                </div>
            </div>
         </ResizablePanel>
      </ResizablePanelGroup>

      <Dialog open={!!toolToConfigure} onOpenChange={(open) => !open && setToolToConfigure(null)}>
        <DialogContent className="sm:max-w-[600px] h-[80vh] flex flex-col p-0 gap-0 overflow-hidden">
            <DialogHeader className="p-6 pb-2">
                <DialogTitle className="flex items-center gap-2 text-xl">
                    <div className="bg-primary/10 p-1.5 rounded-md">
                        <Zap className="w-5 h-5 text-primary" />
                    </div>
                    {toolToConfigure?.name}
                </DialogTitle>
                <DialogDescription>
                    Configure arguments for this tool execution.
                </DialogDescription>
            </DialogHeader>
            <div className="flex-1 overflow-hidden p-6 pt-2">
                {toolToConfigure && (
                    <ToolForm
                        definition={toolToConfigure}
                        onSubmit={handleToolFormSubmit}
                        onCancel={() => setToolToConfigure(null)}
                    />
                )}
            </div>
        </DialogContent>
      </Dialog>

      <Dialog open={!!promptToConfigure} onOpenChange={(open) => !open && setPromptToConfigure(null)}>
        <DialogContent className="sm:max-w-[600px] h-[80vh] flex flex-col p-0 gap-0 overflow-hidden">
            <DialogHeader className="p-6 pb-2">
                <DialogTitle className="flex items-center gap-2 text-xl">
                    <div className="bg-amber-500/10 p-1.5 rounded-md">
                        <MessageSquare className="w-5 h-5 text-amber-500" />
                    </div>
                    {promptToConfigure?.name}
                </DialogTitle>
                <DialogDescription>
                    Configure arguments for this prompt.
                </DialogDescription>
            </DialogHeader>
            <div className="flex-1 overflow-hidden p-6 pt-2">
                {promptToConfigure && (
                    <ToolForm
                        definition={promptToConfigure}
                        onSubmit={handlePromptFormSubmit}
                        onCancel={() => setPromptToConfigure(null)}
                    />
                )}
            </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
