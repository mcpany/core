"use client";

import { useState, useRef, useEffect } from "react";
import { Send, Bot, User, Terminal, Loader2, Sparkles, AlertCircle } from "lucide-react";
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

type MessageType = "user" | "assistant" | "tool-call" | "tool-result" | "error";

interface Message {
  id: string;
  type: MessageType;
  content?: string;
  toolName?: string;
  toolArgs?: any;
  toolResult?: any;
  timestamp: Date;
}

export function PlaygroundClient() {
  const [messages, setMessages] = useState<Message[]>([
      {
          id: "1",
          type: "assistant",
          content: "Hello! I am your MCP Assistant. I can help you interact with your registered tools. Try asking me to list files or analyze something.",
          timestamp: new Date(),
      }
  ]);
  const [input, setInput] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const scrollAreaRef = useRef<HTMLDivElement>(null);

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

    // Mock AI delay
    setTimeout(() => {
        processResponse(input);
    }, 1000);
  };

  const processResponse = (userInput: string) => {
      // Simple mock logic
      const lowerInput = userInput.toLowerCase();

      if (lowerInput.includes("list file") || lowerInput.includes("ls")) {
          // Simulate tool call
          const toolCallMsg: Message = {
              id: Date.now().toString() + "-tool",
              type: "tool-call",
              toolName: "list_files",
              toolArgs: { path: "./" },
              timestamp: new Date(),
          };
          setMessages(prev => [...prev, toolCallMsg]);

          // Simulate tool execution delay
          setTimeout(() => {
              const toolResultMsg: Message = {
                  id: Date.now().toString() + "-result",
                  type: "tool-result",
                  toolName: "list_files",
                  toolResult: { files: ["server.go", "config.yaml", "README.md", "src/"] },
                  timestamp: new Date(),
              };
              setMessages(prev => [...prev, toolResultMsg]);

              // Final assistant response
              setTimeout(() => {
                  setMessages(prev => [...prev, {
                      id: Date.now().toString() + "-final",
                      type: "assistant",
                      content: "I've listed the files for you. There are configuration files and source code present.",
                      timestamp: new Date(),
                  }]);
                  setIsLoading(false);
              }, 800);
          }, 1500);

      } else if (lowerInput.includes("error")) {
           setMessages(prev => [...prev, {
              id: Date.now().toString(),
              type: "error",
              content: "Failed to connect to MCP server: Connection refused.",
              timestamp: new Date(),
          }]);
          setIsLoading(false);
      } else {
          // Generic response
           setMessages(prev => [...prev, {
              id: Date.now().toString(),
              type: "assistant",
              content: "I received your message. I can help you use tools like `list_files` or `read_file`. Just ask!",
              timestamp: new Date(),
          }]);
          setIsLoading(false);
      }
  };

  return (
    <div className="flex flex-col h-full gap-4">
      <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight flex items-center gap-2">
              <Bot className="text-primary" /> Playground
          </h2>
          <Badge variant="outline" className="text-muted-foreground">
              Connected to Localhost
          </Badge>
      </div>

      <Card className="flex-1 flex flex-col overflow-hidden border-muted/50 shadow-sm bg-background/50 backdrop-blur-sm">
        <CardContent className="flex-1 p-0 overflow-hidden flex flex-col">
            <ScrollArea className="flex-1 p-4" ref={scrollAreaRef}>
                <div className="space-y-4 max-w-3xl mx-auto">
                    {messages.map((msg) => (
                        <MessageItem key={msg.id} message={msg} />
                    ))}
                    {isLoading && (
                        <div className="flex items-center gap-2 text-muted-foreground text-sm animate-pulse ml-10">
                            <Sparkles className="size-4" /> Thinking...
                        </div>
                    )}
                </div>
            </ScrollArea>
            <div className="p-4 bg-muted/20 border-t">
                <div className="max-w-3xl mx-auto flex gap-2">
                    <Input
                        placeholder="Type a message to interact with your tools..."
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && handleSend()}
                        disabled={isLoading}
                        className="bg-background shadow-sm"
                    />
                    <Button onClick={handleSend} disabled={isLoading || !input.trim()}>
                        {isLoading ? <Loader2 className="size-4 animate-spin" /> : <Send className="size-4" />}
                        <span className="sr-only">Send</span>
                    </Button>
                </div>
            </div>
        </CardContent>
      </Card>
    </div>
  );
}

function MessageItem({ message }: { message: Message }) {
    if (message.type === "user") {
        return (
            <div className="flex justify-end gap-3">
                 <div className="bg-primary text-primary-foreground rounded-2xl rounded-tr-sm px-4 py-2 max-w-[80%] shadow-md">
                    {message.content}
                </div>
                 <Avatar className="size-8 mt-1 border">
                    <AvatarFallback><User className="size-4" /></AvatarFallback>
                </Avatar>
            </div>
        );
    }

    if (message.type === "assistant") {
        return (
            <div className="flex justify-start gap-3">
                <Avatar className="size-8 mt-1 border bg-muted">
                    <AvatarFallback><Bot className="size-4 text-primary" /></AvatarFallback>
                </Avatar>
                <div className="bg-muted/50 rounded-2xl rounded-tl-sm px-4 py-2 max-w-[80%] shadow-sm border">
                    {message.content}
                </div>
            </div>
        );
    }

    if (message.type === "tool-call") {
        return (
            <div className="flex justify-start gap-3 pl-11 w-full">
                <Card className="w-full max-w-[80%] border-dashed border-primary/30 bg-primary/5">
                    <CardHeader className="p-3 pb-1 flex flex-row items-center gap-2 space-y-0">
                         <Terminal className="size-4 text-primary" />
                         <span className="font-mono text-sm font-medium text-primary">Calling Tool: {message.toolName}</span>
                    </CardHeader>
                    <CardContent className="p-3 pt-1">
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
            <div className="flex justify-start gap-3 pl-11 w-full">
                 <Collapsible className="w-full max-w-[80%] group">
                    <div className="flex items-center justify-between rounded-md border bg-muted/30 px-4 py-2 text-sm shadow-sm">
                        <div className="flex items-center gap-2 text-muted-foreground">
                            <Sparkles className="size-4 text-green-500" />
                            <span>Tool Output ({message.toolName})</span>
                        </div>
                        <CollapsibleTrigger asChild>
                            <Button variant="ghost" size="sm" className="w-9 p-0">
                                <span className="sr-only">Toggle</span>
                                <span className="text-xs underline group-data-[state=open]:no-underline">View</span>
                            </Button>
                        </CollapsibleTrigger>
                    </div>
                    <CollapsibleContent className="mt-2">
                        <pre className="text-xs bg-black text-green-400 p-3 rounded-md border border-green-900/50 font-mono overflow-x-auto">
                            {JSON.stringify(message.toolResult, null, 2)}
                        </pre>
                    </CollapsibleContent>
                </Collapsible>
            </div>
        );
    }

    if (message.type === "error") {
        return (
             <div className="flex justify-start gap-3 pl-11 w-full">
                <div className="flex items-center gap-2 text-destructive bg-destructive/10 px-4 py-2 rounded-md text-sm border border-destructive/20">
                    <AlertCircle className="size-4" />
                    {message.content}
                </div>
            </div>
        );
    }

    return null;
}
