/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
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
  ExternalLink,
  Bug
} from "lucide-react";

import { cn } from "@/lib/utils";
import { apiClient, PromptDefinition } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";

interface PromptWorkbenchProps {
  initialPrompts?: PromptDefinition[];
}

export function PromptWorkbench({ initialPrompts = [] }: PromptWorkbenchProps) {
  const [prompts, setPrompts] = useState<PromptDefinition[]>(initialPrompts);
  const [selectedPrompt, setSelectedPrompt] = useState<PromptDefinition | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [argumentValues, setArgumentValues] = useState<Record<string, string>>({});
  const [executionResult, setExecutionResult] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();
  const { toast } = useToast();

  useEffect(() => {
    if (initialPrompts.length === 0) {
      loadPrompts();
    }
  }, []);

  const loadPrompts = () => {
      apiClient.listPrompts()
          .then((data) => {
              setPrompts(data.prompts || []);
          })
          .catch(err => {
              console.error("Failed to list prompts", err);
              // Silent fail or toast? The list will just be empty.
              toast({
                  title: "Connection Error",
                  description: "Could not fetch prompts from server.",
                  variant: "destructive"
              });
          });
  };

  const loadDemoData = () => {
      setPrompts([
          { name: "summarize_notes", description: "Summarizes meeting notes into key action items", serviceName: "notes-service", arguments: [{name: "notes", description: "The raw notes text", required: true}, {name: "style", description: "bullet or paragraph", required: false}], enabled: true },
          { name: "code_review", description: "Analyzes code for bugs and security issues", serviceName: "dev-service", arguments: [{name: "code", description: "Source code", required: true}, {name: "language", description: "Programming language", required: true}], enabled: true },
          { name: "write_email", description: "Drafts a professional email", serviceName: "office-service", arguments: [{name: "recipient", description: "Name", required: true}, {name: "topic", description: "Main topic", required: true}], enabled: false },
      ]);
  };

  const filteredPrompts = prompts.filter(
    (p) =>
      p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (p.description && p.description.toLowerCase().includes(searchQuery.toLowerCase()))
  );

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
      setExecutionResult(result);

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

  const togglePromptStatus = async (prompt: PromptDefinition) => {
      const newStatus = !prompt.enabled;
      // Optimistic update
      setPrompts(prompts.map(p => p.name === prompt.name ? {...p, enabled: newStatus} : p));
      if (selectedPrompt?.name === prompt.name) {
          setSelectedPrompt({...selectedPrompt, enabled: newStatus});
      }

      try {
          await apiClient.setPromptStatus(prompt.name, newStatus);
      } catch (e) {
          console.error("Failed to toggle status", e);
          toast({
              title: "Error",
              description: "Failed to update prompt status.",
              variant: "destructive"
          });
          // Revert
           setPrompts(prompts.map(p => p.name === prompt.name ? {...p, enabled: !newStatus} : p));
             if (selectedPrompt?.name === prompt.name) {
                setSelectedPrompt({...selectedPrompt, enabled: !newStatus});
            }
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
      // Encode the result or logic to transfer state.
      // Since Playground is a separate page, we might pass data via URL or localStorage.
      // URL might be too long.
      // For now, we'll just navigate to playground.
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
                {filteredPrompts.map((prompt) => (
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
                                {prompt.serviceName || "system"}
                            </Badge>
                             {(prompt.arguments?.length || 0) > 0 && (
                                <span className="text-[10px] text-muted-foreground flex items-center gap-0.5">
                                    <Terminal className="h-3 w-3" /> {prompt.arguments?.length} args
                                </span>
                             )}
                        </div>
                    </button>
                ))}
                {filteredPrompts.length === 0 && (
                    <div className="p-8 text-center text-sm text-muted-foreground flex flex-col items-center gap-2">
                        <p>No prompts found.</p>
                        <Button variant="outline" size="xs" onClick={loadDemoData} className="h-6 text-xs gap-1">
                            <Bug className="h-3 w-3" /> Load Demo Data
                        </Button>
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
                        <div className="flex items-center gap-3">
                             <div className="flex items-center gap-2">
                                <Switch
                                    checked={!!selectedPrompt.enabled}
                                    onCheckedChange={() => togglePromptStatus(selectedPrompt)}
                                />
                                <Label className="text-xs text-muted-foreground">
                                    {selectedPrompt.enabled ? "Enabled" : "Disabled"}
                                </Label>
                             </div>
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
                                        {(selectedPrompt.arguments?.length || 0) > 0 ? (
                                            selectedPrompt.arguments?.map((arg) => (
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
                                            {(executionResult.messages || []).map((msg: any, idx: number) => (
                                                <div key={idx} className="space-y-1">
                                                     <div className="text-[10px] font-mono uppercase text-muted-foreground flex items-center gap-2">
                                                        <span className={cn(
                                                            "w-2 h-2 rounded-full",
                                                            msg.role === "user" ? "bg-blue-500" : "bg-green-500"
                                                        )} />
                                                        {msg.role}
                                                     </div>
                                                     <div className="bg-background border rounded-md p-3 text-sm whitespace-pre-wrap font-mono">
                                                        {msg.content?.type === 'text' ? msg.content.text : typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content)}
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
