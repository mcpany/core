/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import React, { useState, useEffect } from "react";
import { Layers, Cuboid, Plus, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { marketplaceService, ServiceCollection } from "@/lib/marketplace-service";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<ServiceCollection[]>([]);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newStackName, setNewStackName] = useState("");
  const [newStackDesc, setNewStackDesc] = useState("");
  const { toast } = useToast();

  const loadStacks = () => {
      const local = marketplaceService.fetchLocalCollections();
      setStacks(local);
  };

  useEffect(() => {
    loadStacks();
  }, []);

  const handleCreate = () => {
      if (!newStackName.trim()) return;
      const newStack: ServiceCollection = {
          name: newStackName,
          description: newStackDesc,
          author: "User",
          version: "1.0.0",
          services: []
      };
      marketplaceService.saveLocalCollection(newStack);
      toast({ title: "Stack Created", description: `Stack ${newStackName} created successfully.` });
      setIsCreateOpen(false);
      setNewStackName("");
      setNewStackDesc("");
      loadStacks();
  };

  const handleDelete = (e: React.MouseEvent, name: string) => {
      e.preventDefault(); // Prevent Link navigation
      e.stopPropagation();
      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;
      marketplaceService.deleteLocalCollection(name);
      toast({ title: "Stack Deleted", description: "Stack removed successfully." });
      loadStacks();
  };

  return (
    <div className="space-y-6 p-8">
      <div className="flex items-center justify-between">
          <div className="flex flex-col gap-2">
            <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
          </div>
          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
              <DialogTrigger asChild>
                  <Button>
                      <Plus className="mr-2 h-4 w-4" /> Create Stack
                  </Button>
              </DialogTrigger>
              <DialogContent>
                  <DialogHeader>
                      <DialogTitle>Create New Stack</DialogTitle>
                      <DialogDescription>
                          Create a new configuration stack to group your services.
                      </DialogDescription>
                  </DialogHeader>
                  <div className="grid gap-4 py-4">
                      <div className="grid gap-2">
                          <Label htmlFor="name">Name</Label>
                          <Input id="name" value={newStackName} onChange={(e) => setNewStackName(e.target.value)} placeholder="my-stack" />
                      </div>
                      <div className="grid gap-2">
                          <Label htmlFor="desc">Description</Label>
                          <Textarea id="desc" value={newStackDesc} onChange={(e) => setNewStackDesc(e.target.value)} placeholder="Development environment..." />
                      </div>
                  </div>
                  <DialogFooter>
                      <Button onClick={handleCreate} disabled={!newStackName.trim()}>Create</Button>
                  </DialogFooter>
              </DialogContent>
          </Dialog>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {stacks.map((stack) => (
          <Link key={stack.name} href={`/stacks/${stack.name}`}>
             <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full flex flex-col">
               <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    Stack
                  </CardTitle>
                  <div className="flex items-center gap-2">
                      <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                      <button
                        className="text-muted-foreground hover:text-destructive transition-colors opacity-0 group-hover:opacity-100 p-1"
                        onClick={(e) => handleDelete(e, stack.name)}
                        title="Delete Stack"
                      >
                          <Trash2 className="h-4 w-4" />
                      </button>
                  </div>
               </CardHeader>
               <CardContent className="flex-1 flex flex-col justify-between">
                 <div className="flex items-center gap-3 mb-4">
                    <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                        <Layers className="h-6 w-6 text-primary" />
                    </div>
                    <div>
                        <div className="text-2xl font-bold tracking-tight truncate max-w-[150px]" title={stack.name}>{stack.name}</div>
                        <div className="text-xs text-muted-foreground line-clamp-2 h-8">{stack.description || "No description"}</div>
                    </div>
                 </div>

                 <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                    <div className="flex items-center gap-1.5">
                        <span className="relative flex h-2 w-2">
                          <span className="relative inline-flex rounded-full h-2 w-2 bg-slate-400"></span>
                        </span>
                        Local
                    </div>
                    <div>
                        {stack.services.length} Services
                    </div>
                 </div>
               </CardContent>
             </Card>
          </Link>
        ))}
        {stacks.length === 0 && (
            <div className="col-span-full text-center p-12 text-muted-foreground border-2 border-dashed rounded-lg">
                <Layers className="h-12 w-12 mx-auto mb-4 opacity-20" />
                <h3 className="text-lg font-medium">No Stacks Found</h3>
                <p className="mt-2">Create a new stack to get started.</p>
                <Button variant="outline" className="mt-4" onClick={() => setIsCreateOpen(true)}>
                    Create Stack
                </Button>
            </div>
        )}
      </div>
    </div>
  );
}
