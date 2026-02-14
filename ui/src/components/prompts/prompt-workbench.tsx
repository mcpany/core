/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState, useEffect } from "react";
import { apiClient, PromptDefinition } from "@/lib/client";
import { PromptList } from "./prompt-list";
import { PromptEditor } from "./prompt-editor";
import { PromptRunner } from "./prompt-runner";
import { useToast } from "@/hooks/use-toast";
import { MessageSquare } from "lucide-react";

interface PromptWorkbenchProps {
  initialPrompts?: PromptDefinition[];
}

export function PromptWorkbench({ initialPrompts = [] }: PromptWorkbenchProps) {
  const [prompts, setPrompts] = useState<PromptDefinition[]>(initialPrompts);
  const [selectedPrompt, setSelectedPrompt] = useState<PromptDefinition | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    loadPrompts();
  }, []);

  const loadPrompts = async () => {
    try {
        const res = await apiClient.listPrompts();
        // Handle response format from API (might be { prompts: [] } or just [])
        const list = Array.isArray(res) ? res : (res.prompts || []);
        setPrompts(list);
    } catch (e) {
        console.error("Failed to load prompts", e);
        toast({ title: "Error", description: "Failed to load prompts.", variant: "destructive" });
    }
  };

  const handleSelect = (prompt: PromptDefinition) => {
      setSelectedPrompt(prompt);
      setIsEditing(false);
  };

  const handleCreate = () => {
      setSelectedPrompt(null);
      setIsEditing(true);
  };

  const handleSave = async (prompt: PromptDefinition) => {
      try {
          if (selectedPrompt) {
              await apiClient.updatePrompt(prompt);
              toast({ title: "Prompt Updated", description: `Prompt ${prompt.name} saved.` });
          } else {
              await apiClient.createPrompt(prompt);
              toast({ title: "Prompt Created", description: `Prompt ${prompt.name} created.` });
          }
          await loadPrompts();
          // Find the new/updated prompt in the list to select it
          // We fetch the list again to get the server state
          const updatedList = await apiClient.listPrompts();
          const list = Array.isArray(updatedList) ? updatedList : (updatedList.prompts || []);
          const found = list.find((p: any) => p.name === prompt.name);
          if (found) {
              setSelectedPrompt(found);
              setIsEditing(false);
          }
      } catch (e: any) {
          console.error(e);
          // Error handling is done in client.ts throw, caught here
          toast({ title: "Error", description: e.message || "Failed to save prompt.", variant: "destructive" });
      }
  };

  const handleDelete = async (name: string) => {
      try {
          await apiClient.deletePrompt(name);
          toast({ title: "Prompt Deleted", description: `Prompt ${name} deleted.` });
          await loadPrompts();
          setSelectedPrompt(null);
          setIsEditing(false);
      } catch (e: any) {
          console.error(e);
          toast({ title: "Error", description: e.message || "Failed to delete prompt.", variant: "destructive" });
      }
  };

  return (
    <div className="flex h-[calc(100vh-120px)] w-full rounded-md border bg-background shadow-sm overflow-hidden">
      {/* Left Sidebar: Prompt List */}
      <div className="w-[300px] md:w-[350px] shrink-0 h-full">
          <PromptList
            prompts={prompts}
            selectedPrompt={selectedPrompt}
            onSelect={handleSelect}
            onCreate={handleCreate}
          />
      </div>

      {/* Right Pane: Workbench */}
      <div className="flex-1 flex flex-col min-w-0 bg-background h-full">
        {isEditing ? (
            <PromptEditor
                prompt={selectedPrompt}
                onSave={handleSave}
                onDelete={handleDelete}
                onCancel={() => {
                    if (!selectedPrompt) {
                        // Was creating new, go back to empty view or first item?
                        // Just clear selection
                        setIsEditing(false);
                    } else {
                        setIsEditing(false);
                    }
                }}
            />
        ) : selectedPrompt ? (
            <PromptRunner
                prompt={selectedPrompt}
                onEdit={() => setIsEditing(true)}
            />
        ) : (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-8">
                <MessageSquare className="h-12 w-12 opacity-20 mb-4" />
                <h3 className="text-lg font-medium">No Prompt Selected</h3>
                <p className="max-w-xs text-center mt-2">
                    Select a prompt from the library to view details, or create a new one.
                </p>
            </div>
        )}
      </div>
    </div>
  );
}
