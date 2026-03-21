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
  Bug,
  Plus,
  Pencil,
  Trash2
} from "lucide-react";

import { cn } from "@/lib/utils";
import { apiClient, PromptDefinition, UpstreamServiceConfig } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardContent } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { Switch } from "@/components/ui/switch";
import { PromptEditor } from "./prompt-editor";

interface PromptWorkbenchProps {
  initialPrompts?: PromptDefinition[];
}

/**
 * PromptWorkbench is a component that provides an interface for creating, editing, and testing prompts.
 * It allows users to manage prompt templates and test them with different inputs.
 *
 * @param props - The component props.
 * @param props.initialPrompts - The initial list of prompts to display.
 */
export function PromptWorkbench({ initialPrompts = [] }: PromptWorkbenchProps) {
  const [prompts, setPrompts] = useState<PromptDefinition[]>(initialPrompts);
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [selectedPrompt, setSelectedPrompt] = useState<PromptDefinition | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [argumentValues, setArgumentValues] = useState<Record<string, string>>({});
  const [executionResult, setExecutionResult] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Editor State
  const [isEditorOpen, setIsEditorOpen] = useState(false);
  const [editingPrompt, setEditingPrompt] = useState<PromptDefinition | null>(null);

  const router = useRouter();
  const { toast } = useToast();

  useEffect(() => {
    if (initialPrompts.length === 0) {
      loadPrompts();
    }
    loadServices();
  }, []);

  const loadServices = async () => {
      try {
          const list = await apiClient.listServices();
          setServices(list);
      } catch (e) {
          console.error("Failed to load services", e);
      }
  };

  const loadPrompts = () => {
      apiClient.listPrompts()
          .then((data) => {
              // Ensure we have an array
              const list = Array.isArray(data) ? data : (data && Array.isArray(data.prompts) ? data.prompts : []);
              setPrompts(list);
              // Refresh selected prompt if it exists in the new list
              if (selectedPrompt) {
                  const updated = list.find((p: any) => p.name === selectedPrompt.name && (p as any).serviceId === (selectedPrompt as any).serviceId);
                  if (updated) setSelectedPrompt(updated);
              }
          })
          .catch(err => {
              console.warn("Failed to list prompts:", err.message);
              toast({
                  title: "Connection Error",
                  description: "Could not fetch prompts from server.",
                  variant: "destructive"
              });
          });
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

  const getArguments = (prompt: PromptDefinition) => {
      if (!prompt.inputSchema || !prompt.inputSchema.properties) return [];
      const props = prompt.inputSchema.properties as Record<string, any>;
      const required = (prompt.inputSchema.required as string[]) || [];
      return Object.entries(props).map(([key, value]) => ({
          name: key,
          description: value.description,
          required: required.includes(key),
          type: value.type
      }));
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

  const handleSavePrompt = async (newPrompt: PromptDefinition, serviceId: string) => {
      try {
          const service = await apiClient.getService(serviceId);
          if (!service) throw new Error("Service not found");

          // Ensure prompts array exists
          const currentPrompts = service.service.prompts || [];

          // Check if updating existing
          const existingIdx = currentPrompts.findIndex((p: any) => p.name === newPrompt.name);

          let updatedPrompts = [...currentPrompts];
          if (existingIdx >= 0) {
              updatedPrompts[existingIdx] = newPrompt;
          } else {
              updatedPrompts.push(newPrompt);
          }

          const updatedService = {
              ...service.service,
              prompts: updatedPrompts
          };

          await apiClient.updateService(updatedService);
          toast({ title: "Prompt Saved", description: `Prompt ${newPrompt.name} saved to service ${service.service.name}` });

          loadPrompts();
          setIsEditorOpen(false);
      } catch (e) {
          console.error(e);
          toast({ title: "Failed to save prompt", description: String(e), variant: "destructive" });
      }
  };

  const handleDeletePrompt = async (prompt: PromptDefinition) => {
      if (!confirm(`Are you sure you want to delete prompt "${prompt.name}"?`)) return;

      const serviceId = (prompt as any).serviceId;
      if (!serviceId) {
          toast({ title: "Error", description: "Cannot delete prompt: Service ID unknown", variant: "destructive" });
          return;
      }

      try {
          const service = await apiClient.getService(serviceId);
          if (!service) throw new Error("Service not found");

          const updatedPrompts = (service.service.prompts || []).filter((p: any) => p.name !== prompt.name);

          const updatedService = {
              ...service.service,
              prompts: updatedPrompts
          };

          await apiClient.updateService(updatedService);
          toast({ title: "Prompt Deleted", description: `Prompt ${prompt.name} removed.` });

          if (selectedPrompt?.name === prompt.name) {
              setSelectedPrompt(null);
          }
          loadPrompts();
      } catch (e) {
          console.error(e);
          toast({ title: "Failed to delete prompt", description: String(e), variant: "destructive" });
      }
  };

  const togglePromptStatus = async (prompt: PromptDefinition) => {
      const newDisable = !prompt.disable;
      // Optimistic update
      setPrompts(prompts.map(p => p.name === prompt.name ? {...p, disable: newDisable} : p));
      if (selectedPrompt?.name === prompt.name) {
          setSelectedPrompt({...selectedPrompt, disable: newDisable});
      }

      try {
          await apiClient.setPromptStatus(prompt.name, newDisable);
      } catch (e) {
          console.error("Failed to toggle status", e);
          toast({ title: "Error", description: "Failed to update prompt status.", variant: "destructive" });
          loadPrompts(); // Revert
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
            <div className="flex items-center justify-between">
                <h3 className="font-semibold text-sm flex items-center gap-2">
                    <MessageSquare className="h-4 w-4" /> Prompt Library
                </h3>
                <Button size="icon" variant="ghost" className="h-6 w-6" onClick={() => {
                    setEditingPrompt(null);
                    setIsEditorOpen(true);
                }}>
                    <Plus className="h-4 w-4" />
                </Button>
            </div>
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
                                {(prompt as any).serviceId || "System"}
                            </Badge>
                             {(getArguments(prompt).length || 0) > 0 && (
                                <span className="text-[10px] text-muted-foreground flex items-center gap-0.5">
                                    <Terminal className="h-3 w-3" /> {getArguments(prompt).length} args
                                </span>
                             )}
                        </div>
                    </button>
                ))}
                {filteredPrompts.length === 0 && (
                    <div className="p-8 text-center text-sm text-muted-foreground flex flex-col items-center gap-2">
                        <p>No prompts found.</p>
                        <Button variant="outline" size="sm" onClick={() => setIsEditorOpen(true)} className="h-6 text-xs gap-1">
                            <Plus className="h-3 w-3" /> Create First Prompt
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
                            <h2 className="text-2xl font-bold tracking-tight flex items-center gap-2">
                                {selectedPrompt.name}
                                <Button size="icon" variant="ghost" className="h-6 w-6 opacity-50 hover:opacity-100" onClick={() => {
                                    setEditingPrompt(selectedPrompt);
                                    setIsEditorOpen(true);
                                }}>
                                    <Pencil className="h-3 w-3" />
                                </Button>
                            </h2>
                            <p className="text-muted-foreground mt-1">{selectedPrompt.description || "No description provided."}</p>
                        </div>
                        <div className="flex items-center gap-3">
                             <div className="flex items-center gap-2">
                                <Switch
                                    checked={!selectedPrompt.disable}
                                    onCheckedChange={() => togglePromptStatus(selectedPrompt)}
                                />
                                <Label className="text-xs text-muted-foreground">
                                    {!selectedPrompt.disable ? "Enabled" : "Disabled"}
                                </Label>
                             </div>
                             <Button variant="destructive" size="icon" className="h-8 w-8" onClick={() => handleDeletePrompt(selectedPrompt)}>
                                 <Trash2 className="h-4 w-4" />
                             </Button>
                        </div>
                    </div>
                </div>

                {/* Content */}
                <div className="flex-1 overflow-y-auto p-6" data-testid="prompt-details">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 h-full">
                        {/* Arguments Column */}
                        <div className="flex flex-col gap-6">
                             <div>
                                <h3 className="text-sm font-medium mb-4 flex items-center gap-2 text-primary">
                                    <Terminal className="h-4 w-4" /> Configuration
                                </h3>
                                <Card>
                                    <CardContent className="p-4 space-y-4">
                                        {getArguments(selectedPrompt).length > 0 ? (
                                            getArguments(selectedPrompt).map((arg) => (
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
                                            {(executionResult?.messages || []).map((msg: any, idx: number) => (
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
                <Button variant="outline" onClick={() => setIsEditorOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" /> Create New Prompt
                </Button>
            </div>
        )}

        <PromptEditor
            open={isEditorOpen}
            onOpenChange={setIsEditorOpen}
            prompt={editingPrompt}
            services={services}
            onSave={handleSavePrompt}
        />
      </div>
    </div>
  );
}
