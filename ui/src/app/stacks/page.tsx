/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { StackCard, Stack } from "@/components/stacks/stack-card";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";
import { Layers } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchStacks = async () => {
    setLoading(true);
    try {
        const res = await apiClient.listCollections();
        setStacks(res || []);
    } catch (e) {
        console.error(e);
        toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to load stacks."
        });
    } finally {
        setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const handleDelete = async (name: string) => {
    try {
        await apiClient.deleteCollection(name);
        setStacks(prev => prev.filter(s => s.name !== name));
        toast({
            title: "Stack Deleted",
            description: `Stack ${name} has been removed.`
        });
    } catch (e) {
         console.error(e);
         // Refresh to ensure sync
         fetchStacks();
         toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to delete stack."
        });
    }
  };

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <CreateStackDialog onStackCreated={fetchStacks} />
      </div>

      {loading ? (
           <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
               {[...Array(3)].map((_, i) => (
                   <div key={i} className="h-40 rounded-lg bg-muted/20 animate-pulse" />
               ))}
           </div>
      ) : stacks.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-muted/10">
               <Layers className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
               <h3 className="text-lg font-medium">No stacks found</h3>
               <p className="text-sm text-muted-foreground mb-4">Create your first stack to organize services.</p>
               <CreateStackDialog onStackCreated={fetchStacks} />
          </div>
      ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
              <StackCard key={stack.name} stack={stack} onDelete={handleDelete} />
            ))}
          </div>
      )}
    </div>
  );
}
