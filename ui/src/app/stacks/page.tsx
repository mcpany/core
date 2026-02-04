/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Layers, Cuboid, Trash2, Loader2, Activity } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";

interface Stack {
    name: string;
    description?: string;
    services?: any[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();
  const router = useRouter();

  const fetchStacks = async () => {
    setLoading(true);
    try {
        const res = await apiClient.listCollections();
        setStacks(Array.isArray(res) ? res : res.collections || []);
    } catch (e) {
        console.error("Failed to load stacks", e);
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

  const handleCreate = async (name: string) => {
      try {
          // Check if exists
          if (stacks.some(s => s.name === name)) {
              toast({
                  variant: "destructive",
                  title: "Error",
                  description: "Stack already exists."
              });
              return;
          }

          // Create empty stack
          const newStack = {
              name: name,
              services: []
          };

          await apiClient.saveCollection(newStack);

          toast({
              title: "Stack Created",
              description: `Stack ${name} has been created.`
          });

          router.push(`/stacks/${name}`);
      } catch (e) {
          console.error("Failed to create stack", e);
          throw e; // Let dialog handle loading state reset via try/catch
      }
  };

  const handleDelete = async (e: React.MouseEvent, name: string) => {
      e.preventDefault(); // Prevent navigation
      e.stopPropagation();

      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

      try {
          await apiClient.deleteCollection(name);
          setStacks(prev => prev.filter(s => s.name !== name));
          toast({
              title: "Stack Deleted",
              description: `Stack ${name} has been removed.`
          });
      } catch (err) {
          console.error("Failed to delete stack", err);
          toast({
              variant: "destructive",
              title: "Error",
              description: "Failed to delete stack."
          });
      }
  };

  return (
    <div className="space-y-6 p-8 h-[calc(100vh-4rem)] overflow-y-auto">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <CreateStackDialog onCreate={handleCreate} />
      </div>

      {loading ? (
          <div className="flex justify-center items-center h-64">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.length === 0 && (
                <div className="col-span-full flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-lg bg-muted/20">
                    <Layers className="h-12 w-12 text-muted-foreground mb-4 opacity-50" />
                    <h3 className="text-lg font-medium">No stacks found</h3>
                    <p className="text-sm text-muted-foreground mb-4">Create your first stack to get started.</p>
                </div>
            )}

            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
                <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 stack-card">
                <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                        Stack
                    </CardTitle>
                    <div className="flex items-center gap-2">
                         <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity text-destructive hover:text-destructive hover:bg-destructive/10"
                            onClick={(e) => handleDelete(e, stack.name)}
                            title="Delete Stack"
                         >
                            <Trash2 className="h-4 w-4" />
                         </Button>
                         <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center gap-3 mb-4">
                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                            <Layers className="h-6 w-6 text-primary" />
                        </div>
                        <div>
                            <div className="text-2xl font-bold tracking-tight truncate max-w-[200px]" title={stack.name}>{stack.name}</div>
                            <div className="text-xs text-muted-foreground font-mono truncate max-w-[200px]">{stack.description || "No description"}</div>
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            <Activity className="h-3 w-3" />
                            Active
                        </div>
                        <div>
                            {stack.services?.length || 0} Services
                        </div>
                    </div>
                </CardContent>
                </Card>
            </Link>
            ))}
        </div>
      )}
    </div>
  );
}
