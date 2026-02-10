/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useRouter } from "next/navigation";
import { useToast } from "@/hooks/use-toast";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [newStackName, setNewStackName] = useState("");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const router = useRouter();
  const { toast } = useToast();

  useEffect(() => {
    fetchStacks();
  }, []);

  const fetchStacks = async () => {
    try {
      setLoading(true);
      const res = await apiClient.listCollections();
      setStacks(res || []);
    } catch (e) {
      console.error("Failed to fetch stacks", e);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateStack = async () => {
      if (!newStackName.trim()) return;

      try {
          // Create an empty stack (collection)
          await apiClient.saveCollection({
              name: newStackName,
              services: []
          });
          setIsDialogOpen(false);
          setNewStackName("");
          router.push(`/stacks/${newStackName}`);
      } catch (e) {
          console.error("Failed to create stack", e);
          toast({
              title: "Error",
              description: "Failed to create stack.",
              variant: "destructive"
          });
      }
  };

  const handleDeleteStack = async (e: React.MouseEvent, name: string) => {
      e.preventDefault(); // Prevent navigation
      e.stopPropagation();

      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

      try {
          await apiClient.deleteCollection(name);
          fetchStacks();
          toast({
              title: "Stack Deleted",
              description: `Stack ${name} has been removed.`
          });
      } catch (err) {
          console.error("Failed to delete stack", err);
          toast({
              title: "Error",
              description: "Failed to delete stack.",
              variant: "destructive"
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
                        Enter a name for your new configuration stack.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="name" className="text-right">
                            Name
                        </Label>
                        <Input
                            id="name"
                            value={newStackName}
                            onChange={(e) => setNewStackName(e.target.value)}
                            className="col-span-3"
                            placeholder="my-mcp-stack"
                        />
                    </div>
                </div>
                <DialogFooter>
                    <Button onClick={handleCreateStack}>Create</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
      </div>

      {loading ? (
           <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
               {[1, 2, 3].map(i => (
                   <div key={i} className="h-40 rounded-xl bg-muted animate-pulse" />
               ))}
           </div>
      ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
              <Link key={stack.name} href={`/stacks/${stack.name}`}>
                 <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 relative">
                   <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                      <CardTitle className="text-sm font-medium text-muted-foreground">
                        Stack
                      </CardTitle>
                      <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                   </CardHeader>
                   <CardContent>
                     <div className="flex items-center gap-3 mb-4">
                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                            <Layers className="h-6 w-6 text-primary" />
                        </div>
                        <div>
                            <div className="text-xl font-bold tracking-tight truncate max-w-[150px]" title={stack.name}>{stack.name}</div>
                            <div className="text-xs text-muted-foreground font-mono truncate max-w-[150px]">{stack.description || "No description"}</div>
                        </div>
                     </div>

                     <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            <span className="relative flex h-2 w-2">
                              <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                            </span>
                            Active
                        </div>
                        <div>
                            {stack.services?.length || 0} Services
                        </div>
                     </div>
                   </CardContent>
                   <Button
                        variant="ghost"
                        size="icon"
                        className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity text-destructive hover:text-destructive hover:bg-destructive/10 h-6 w-6"
                        onClick={(e) => handleDeleteStack(e, stack.name)}
                   >
                       <Trash2 className="h-3 w-3" />
                   </Button>
                 </Card>
              </Link>
            ))}
            {stacks.length === 0 && (
                <div className="col-span-full text-center py-20 border-2 border-dashed rounded-xl text-muted-foreground">
                    No stacks found. Create one to get started.
                </div>
            )}
          </div>
      )}
    </div>
  );
}
