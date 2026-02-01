/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, Trash2, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

interface Stack {
  name: string;
  id?: string;
  services?: any[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newStackName, setNewStackName] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const { toast } = useToast();

  const fetchStacks = async (showLoading = true) => {
    if (showLoading) setLoading(true);
    try {
      const res = await apiClient.listCollections();
      // Handle array or object wrapper
      if (Array.isArray(res)) {
        setStacks(res);
      } else if (res && Array.isArray(res.collections)) {
        setStacks(res.collections);
      } else {
        setStacks([]);
      }
    } catch (e) {
      console.error("Failed to list stacks", e);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to load stacks."
      });
    } finally {
      if (showLoading) setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const handleCreate = async () => {
    if (!newStackName.trim()) return;

    setIsCreating(true);
    try {
      await apiClient.saveCollection({
        name: newStackName,
        services: []
      });
      toast({
        title: "Stack Created",
        description: `Stack "${newStackName}" has been created.`
      });
      setIsDialogOpen(false);
      setNewStackName("");
      fetchStacks(false); // Background refresh
    } catch (e: any) {
      console.error("Failed to create stack", e);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to create stack: " + (e.message || "Unknown error")
      });
    } finally {
      setIsCreating(false);
    }
  };

  const handleDelete = async (e: React.MouseEvent, name: string) => {
    e.preventDefault(); // Prevent navigation
    e.stopPropagation();

    if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

    // Optimistic update
    const previousStacks = [...stacks];
    setStacks(prev => prev.filter(s => s.name !== name));

    try {
      await apiClient.deleteCollection(name);
      toast({
        title: "Stack Deleted",
        description: `Stack "${name}" has been deleted.`
      });
      fetchStacks(false); // Background refresh to ensure sync
    } catch (error: any) {
      console.error("Failed to delete stack", error);
      setStacks(previousStacks); // Revert
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
          <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
          <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" /> Create Stack
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Stack</DialogTitle>
              <DialogDescription>
                Enter a name for your new service collection.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Input
                  id="name"
                  placeholder="e.g. production-stack"
                  value={newStackName}
                  onChange={(e) => setNewStackName(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleCreate()}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
              <Button onClick={handleCreate} disabled={isCreating || !newStackName.trim()}>
                {isCreating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {loading && stacks.length === 0 ? (
        <div className="flex items-center justify-center h-40 text-muted-foreground">
          <Loader2 className="h-8 w-8 animate-spin mr-2" /> Loading stacks...
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
               <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full flex flex-col">
                 <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                      Stack
                    </CardTitle>
                    <div className="flex gap-2">
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 text-destructive opacity-0 group-hover:opacity-100 transition-opacity"
                            onClick={(e) => handleDelete(e, stack.name)}
                            title="Delete Stack"
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                        <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                    </div>
                 </CardHeader>
                 <CardContent className="flex-1 flex flex-col">
                   <div className="flex items-center gap-3 mb-4 flex-1">
                      <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform h-fit">
                          <Layers className="h-6 w-6 text-primary" />
                      </div>
                      <div className="min-w-0">
                          <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                          <div className="text-xs text-muted-foreground font-mono truncate">{stack.id || stack.name}</div>
                      </div>
                   </div>

                   <div className="flex items-center justify-between text-xs text-muted-foreground mt-auto pt-4 border-t">
                      <div className="flex items-center gap-1.5">
                          <span className="relative flex h-2 w-2">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                          </span>
                          Online
                      </div>
                      <div>
                          {(stack.services || []).length} Services
                      </div>
                   </div>
                 </CardContent>
               </Card>
            </Link>
          ))}
          {stacks.length === 0 && !loading && (
              <div className="col-span-full text-center py-12 border-2 border-dashed rounded-lg text-muted-foreground">
                  <Layers className="h-10 w-10 mx-auto mb-2 opacity-20" />
                  <p>No stacks found.</p>
                  <Button variant="link" onClick={() => setIsDialogOpen(true)}>Create one</Button>
              </div>
          )}
        </div>
      )}
    </div>
  );
}
