/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import {
  MessageSquare,
  Search,
  Play,
  Copy,
  Terminal,
  ChevronRight,
  Sparkles,
  Loader2,
  ExternalLink
} from "lucide-react";

import { cn } from "@/lib/utils";
import { apiClient, PromptDefinition } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardContent } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { useToast } from "@/hooks/use-toast";

interface PromptWorkbenchProps {
  initialPrompts?: PromptDefinition[];
}

interface ParsedArgument {
    name: string;
    description?: string;
    required: boolean;
}

// Define specific type for Execution Result
interface ExecutionResult {
    messages: {
        role: string;
        content: string | { type: string; text: string };
    }[];
}

const MOCK_PROMPTS: PromptDefinition[] = [
    {
        name: "code_review",
        title: "Code Review",
        description: "Analyze code for bugs and security issues",
        disable: false,
        inputSchema: {
            type: "object",
            properties: {
                code: { type: "string", description: "The source code to review" },
                language: { type: "string", description: "The programming language" }
            },
            required: ["code"]
        },
        messages: [],
        profiles: []
    },
    {
        name: "summarize_text",
        title: "Summarize Text",
        description: "Create a concise summary of the provided text",
        disable: false,
        inputSchema: {
            type: "object",
            properties: {
                text: { type: "string", description: "Text to summarize" },
                length: { type: "string", description: "short, medium, or long" }
            }
        },
        messages: [],
        profiles: []
    },
    {
        name: "git_commit_msg",
        title: "Git Commit Message",
        description: "Generate a conventional commit message",
        disable: true,
        inputSchema: {
             type: "object",
             properties: {
                 diff: { type: "string", description: "Git diff output" }
             },
             required: ["diff"]
        },
        messages: [],
        profiles: []
    }
];

function parseArgumentsFromSchema(schema: Record<string, unknown> | undefined): ParsedArgument[] {
    if (!schema || !schema.properties) return [];

    // Explicitly cast to Record<string, unknown> if we trust the structure or use type guards
    const props = schema.properties as Record<string, { description?: string }>;
    const requiredSet = new Set(Array.isArray(schema.required) ? (schema.required as string[]) : []);

    return Object.entries(props).map(([key, value]) => ({
        name: key,
        description: value.description || "",
        required: requiredSet.has(key)
    }));
}

export function PromptWorkbench({ initialPrompts = [] }: PromptWorkbenchProps) {
  const [prompts, setPrompts] = useState<PromptDefinition[]>(initialPrompts);
  const [selectedPrompt, setSelectedPrompt] = useState<PromptDefinition | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [argumentValues, setArgumentValues] = useState<Record<string, string>>({});
  const [executionResult, setExecutionResult] = useState<ExecutionResult | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();
  const { toast } = useToast();

  useEffect(() => {
    if (initialPrompts.length === 0) {
      apiClient.listPrompts()
        .then((data) => {
            setPrompts(data.prompts || []);
        })
        .catch((e) => {
            console.warn("Failed to load prompts, using mock data for demo", e);
            setPrompts(MOCK_PROMPTS);
            toast({
                title: "Demo Mode",
                description: "Using mock prompts because the backend is unavailable.",
                variant: "default"
            });
        });
    }
  }, []);

  const filteredPrompts = prompts.filter(
    (p) =>
      p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (p.description && p.description.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  const parsedArgs = useMemo(() => {
      if (!selectedPrompt) return [];
      // Cast to Record<string, unknown> safely as inputSchema is Struct
      return parseArgumentsFromSchema(selectedPrompt.inputSchema as Record<string, unknown>);
  }, [selectedPrompt]);

  const handleSelectPrompt = (prompt: PromptDefinition) => {
    setSelectedPrompt(prompt);
    setArgumentValues({});
    setExecutionResult(null);
  };

  const handleExecute = async () => {
    if (!selectedPrompt) return;

    setIsLoading(true);
    try {
      const result = await apiClient.executePrompt(selectedPrompt.name, argumentValues);
      setExecutionResult(result as ExecutionResult);

      toast({
        title: "Prompt Executed",
        description: "Successfully generated prompt messages.",
      });
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        title: "Execution Failed",
        description: "Failed to execute prompt.",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const togglePromptStatus = async (enabled: boolean) => {
      if (!selectedPrompt) return;

      // Optimistic update
      const updatedPrompt = { ...selectedPrompt, disable: !enabled };
      setSelectedPrompt(updatedPrompt);
      setPrompts(prompts.map(p => p.name === selectedPrompt.name ? updatedPrompt : p));

      try {
          await apiClient.setPromptStatus(selectedPrompt.name, enabled);
          toast({
              title: enabled ? "Prompt Enabled" : "Prompt Disabled",
              description: `Prompt '${selectedPrompt.name}' status updated.`
          });
      } catch (error) {
          console.error(error);
          // Revert
          setSelectedPrompt(selectedPrompt);
          setPrompts(prompts.map(p => p.name === selectedPrompt.name ? selectedPrompt : p));
          toast({
              variant: "destructive",
              title: "Update Failed",
              description: "Failed to update prompt status."
          });
      }
  };

  const copyToClipboard = () => {
    if (executionResult) {
      navigator.clipboard.writeText(JSON.stringify(executionResult, null, 2));
      toast({
        title: "Copied",
        description: "Result copied to clipboard.",
      });
    }
  };

  const openInPlayground = () => {
      router.push("/playground");
      toast({
          title: "Navigating to Playground",
          description: "You can paste the prompt result there.",
      });
  };

  return (
    <div className="flex h-[calc(100vh-120px)] w-full rounded-md border bg-background shadow-sm overflow-hidden">
      {/* Left Sidebar: Prompt List */}
      <div className="w-[300px] md:w-[350px] border-r flex flex-col bg-muted/10">
        <div className="p-4 border-b space-y-3">
            <h3 className="font-semibold text-sm flex items-center gap-2">
                <MessageSquare className="h-4 w-4" /> Prompt Library
            </h3>
            <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search prompts..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-8 h-9 text-sm"
                />
            </div>
        </div>
        <ScrollArea className="flex-1">
            <div className="flex flex-col p-2 gap-1">
                {filteredPrompts.map((prompt) => {
                    // Calculate arg count efficiently
                    const argCount = parseArgumentsFromSchema(prompt.inputSchema as Record<string, unknown>).length;
                    return (
                        <button
                            key={prompt.name}
                            onClick={() => handleSelectPrompt(prompt)}
                            className={cn(
                                "flex flex-col items-start gap-1 p-3 rounded-md text-left transition-colors hover:bg-accent hover:text-accent-foreground",
                                selectedPrompt?.name === prompt.name ? "bg-accent text-accent-foreground shadow-sm" : ""
                            )}
                        >
                            <div className="flex items-center justify-between w-full">
                                <span className="font-medium text-sm truncate">{prompt.name}</span>
                                {selectedPrompt?.name === prompt.name && <ChevronRight className="h-3 w-3 opacity-50" />}
                            </div>
                            {prompt.description && (
                                <p className="text-xs text-muted-foreground line-clamp-2">
                                    {prompt.description}
                                </p>
                            )}
                            <div className="flex items-center gap-2 mt-1">
                                <Badge variant="outline" className="text-[10px] px-1 py-0 h-4">
                                    System
                                </Badge>
                                 {argCount > 0 && (
                                    <span className="text-[10px] text-muted-foreground flex items-center gap-0.5">
                                        <Terminal className="h-3 w-3" /> {argCount} args
                                    </span>
                                 )}
                            </div>
                        </button>
                    );
                })}
                {filteredPrompts.length === 0 && (
                    <div className="p-4 text-center text-sm text-muted-foreground">
                        No prompts found.
                    </div>
                )}
            </div>
        </ScrollArea>
      </div>

      {/* Right Pane: Workbench */}
      <div className="flex-1 flex flex-col min-w-0 bg-background">
        {selectedPrompt ? (
            <div className="flex flex-col h-full">
                {/* Header */}
                <div className="p-6 border-b pb-4">
                    <div className="flex items-start justify-between">
                        <div>
                            <h2 className="text-2xl font-bold tracking-tight">{selectedPrompt.name}</h2>
                            <p className="text-muted-foreground mt-1">{selectedPrompt.description || "No description provided."}</p>
                        </div>
                        <div className="flex items-center gap-2">
                            <Label htmlFor="prompt-status" className="text-sm text-muted-foreground">
                                {selectedPrompt.disable ? "Disabled" : "Enabled"}
                            </Label>
                            <Switch
                                id="prompt-status"
                                checked={!selectedPrompt.disable}
                                onCheckedChange={togglePromptStatus}
                            />
                        </div>
                    </div>
                </div>

                {/* Content */}
                <div className="flex-1 overflow-y-auto p-6">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 h-full">
                        {/* Arguments Column */}
                        <div className="flex flex-col gap-6">
                             <div>
                                <h3 className="text-sm font-medium mb-4 flex items-center gap-2 text-primary">
                                    <Terminal className="h-4 w-4" /> Configuration
                                </h3>
                                <Card>
                                    <CardContent className="p-4 space-y-4">
                                        {parsedArgs.length > 0 ? (
                                            parsedArgs.map((arg) => (
                                                <div key={arg.name} className="space-y-1.5">
                                                    <Label htmlFor={arg.name} className="flex items-center gap-1 text-xs font-mono uppercase text-muted-foreground">
                                                        {arg.name}
                                                        {arg.required && <span className="text-red-500">*</span>}
                                                    </Label>
                                                    <Input
                                                        id={arg.name}
                                                        placeholder={arg.description}
                                                        value={argumentValues[arg.name] || ""}
                                                        onChange={(e) => setArgumentValues({...argumentValues, [arg.name]: e.target.value})}
                                                    />
                                                    {arg.description && <p className="text-[10px] text-muted-foreground">{arg.description}</p>}
                                                </div>
                                            ))
                                        ) : (
                                            <div className="text-sm text-muted-foreground italic text-center py-4">
                                                No arguments required.
                                            </div>
                                        )}
                                        <Button
                                            className="w-full mt-4"
                                            onClick={handleExecute}
                                            disabled={isLoading}
                                        >
                                            {isLoading ? (
                                                <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Generating...</>
                                            ) : (
                                                <><Play className="mr-2 h-4 w-4" /> Generate Preview</>
                                            )}
                                        </Button>
                                    </CardContent>
                                </Card>
                            </div>
                        </div>

                        {/* Preview Column */}
                        <div className="flex flex-col h-full min-h-[400px]">
                            <div className="flex items-center justify-between mb-4">
                                <h3 className="text-sm font-medium flex items-center gap-2 text-primary">
                                    <Sparkles className="h-4 w-4" /> Output Preview
                                </h3>
                                <div className="flex items-center gap-2">
                                     <Button variant="ghost" size="sm" onClick={copyToClipboard} disabled={!executionResult}>
                                        <Copy className="h-3 w-3" />
                                     </Button>
                                </div>
                            </div>
                            <Card className="flex-1 flex flex-col overflow-hidden bg-muted/30 border-dashed">
                                <CardContent className="flex-1 p-0 overflow-auto">
                                    {executionResult ? (
                                        <div className="p-4 space-y-4">
                                            {(executionResult.messages || []).map((msg, idx) => (
                                                <div key={idx} className="space-y-1">
                                                     <div className="text-[10px] font-mono uppercase text-muted-foreground flex items-center gap-2">
                                                        <span className={cn(
                                                            "w-2 h-2 rounded-full",
                                                            msg.role === "user" ? "bg-blue-500" : "bg-green-500"
                                                        )} />
                                                        {msg.role}
                                                     </div>
                                                     <div className="bg-background border rounded-md p-3 text-sm whitespace-pre-wrap font-mono">
                                                        {/* Handle string or object content safely */}
                                                        {typeof msg.content === 'string'
                                                            ? msg.content
                                                            : (msg.content as { text?: string }).text || JSON.stringify(msg.content)
                                                        }
                                                     </div>
                                                </div>
                                            ))}
                                            <div className="pt-4 flex justify-end">
                                                <Button size="sm" variant="outline" onClick={openInPlayground}>
                                                    Open in Playground <ExternalLink className="ml-2 h-3 w-3" />
                                                </Button>
                                            </div>
                                        </div>
                                    ) : (
                                        <div className="flex flex-col items-center justify-center h-full text-muted-foreground text-sm p-8 text-center">
                                            <Sparkles className="h-10 w-10 opacity-20 mb-3" />
                                            <p>Configure arguments and click Generate to see the prompt result.</p>
                                        </div>
                                    )}
                                </CardContent>
                            </Card>
                        </div>
                    </div>
                </div>
            </div>
        ) : (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-8">
                <MessageSquare className="h-12 w-12 opacity-20 mb-4" />
                <h3 className="text-lg font-medium">No Prompt Selected</h3>
                <p className="max-w-xs text-center mt-2">
                    Select a prompt from the library to view details, configure arguments, and test execution.
                </p>
            </div>
        )}
      </div>
    </div>
  );
}
