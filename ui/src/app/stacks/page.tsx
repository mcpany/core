/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Layers, Cuboid, Plus, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newStackName, setNewStackName] = useState("");
  const [creating, setCreating] = useState(false);
  const { toast } = useToast();
  const router = useRouter();

  useEffect(() => {
    fetchStacks();
  }, []);

  const fetchStacks = async () => {
    setLoading(true);
    try {
      const data = await apiClient.listCollections();
      setStacks(data);
    } catch (e) {
      console.error("Failed to list stacks", e);
      toast({
          title: "Error",
          description: "Failed to load stacks.",
          variant: "destructive"
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
      if (!newStackName.trim()) return;
      setCreating(true);
      try {
          // Create empty collection
          await apiClient.saveCollection({
              name: newStackName,
              version: "1.0.0",
              description: "New stack",
              services: []
          });
          toast({
              title: "Stack Created",
              description: `Stack ${newStackName} created successfully.`
          });
          setIsCreateOpen(false);
          setNewStackName("");
          // Redirect to editor
          router.push(`/stacks/${newStackName}`);
      } catch (e) {
          console.error(e);
          toast({
              title: "Error",
              description: "Failed to create stack.",
              variant: "destructive"
          });
      } finally {
          setCreating(false);
      }
  };

  const handleDelete = async (name: string, e: React.MouseEvent) => {
      e.preventDefault(); // Prevent navigation
      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

      try {
          await apiClient.deleteCollection(name);
          toast({
              title: "Stack Deleted",
              description: `Stack ${name} has been removed.`
          });
          fetchStacks();
      } catch (e) {
          console.error(e);
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
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> New Stack
                </Button>
            </DialogTrigger>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Create New Stack</DialogTitle>
                    <DialogDescription>
                        Enter a unique name for your new stack.
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
                            placeholder="my-stack"
                        />
                    </div>
                </div>
                <DialogFooter>
                    <Button onClick={handleCreate} disabled={creating || !newStackName}>
                        {creating ? "Creating..." : "Create"}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
      </div>

      {loading ? (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map(i => (
                  <div key={i} className="h-48 rounded-xl bg-muted/20 animate-pulse" />
              ))}
          </div>
      ) : stacks.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 border-2 border-dashed rounded-lg bg-muted/10">
              <Layers className="h-12 w-12 text-muted-foreground opacity-20 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground">No stacks found</h3>
              <p className="text-sm text-muted-foreground mb-4">Create your first stack to get started.</p>
              <Button onClick={() => setIsCreateOpen(true)} variant="outline">Create Stack</Button>
          </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
                <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full flex flex-col">
                <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                    <CardTitle className="text-sm font-medium text-muted-foreground truncate w-full pr-4">
                        {stack.description || "No description"}
                    </CardTitle>
                    <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors shrink-0" />
                </CardHeader>
                <CardContent className="flex-1">
                    <div className="flex items-center gap-3 mb-4">
                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                            <Layers className="h-6 w-6 text-primary" />
                        </div>
                        <div className="min-w-0">
                            <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                            <div className="text-xs text-muted-foreground font-mono">{stack.version}</div>
                        </div>
                    </div>
                </CardContent>
                <CardFooter className="border-t pt-4 text-xs text-muted-foreground flex justify-between items-center bg-muted/20">
                     <div className="flex items-center gap-1.5">
                         <Badge variant="secondary" className="text-[10px] h-5">
                             {stack.services ? stack.services.length : 0} Services
                         </Badge>
                     </div>
                     <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6 hover:bg-destructive/10 hover:text-destructive"
                        onClick={(e) => handleDelete(stack.name, e)}
                        title="Delete Stack"
                     >
                         <Trash2 className="h-3 w-3" />
                     </Button>
                </CardFooter>
                </Card>
            </Link>
            ))}
        </div>
      )}
    </div>
  );
}
