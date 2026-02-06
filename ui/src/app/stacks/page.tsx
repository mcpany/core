/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Layers, Cuboid, Plus, Loader2, Trash2, AlertTriangle } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";

interface Stack {
    name: string;
    version?: string;
    services?: any[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);

  // Create State
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newStackName, setNewStackName] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  // Delete State
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [stackToDelete, setStackToDelete] = useState<string | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const { toast } = useToast();
  const router = useRouter();

  const fetchStacks = async () => {
    setLoading(true);
    try {
        const res = await apiClient.listCollections();
        setStacks(res || []);
    } catch (error) {
        console.error("Failed to fetch stacks", error);
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

  const handleCreate = async () => {
      if (!newStackName.trim()) {
          toast({
              variant: "destructive",
              title: "Validation Error",
              description: "Stack name is required."
          });
          return;
      }

      setIsCreating(true);
      try {
          // Create an empty stack (collection)
          await apiClient.saveCollection({
              name: newStackName,
              services: []
          });
          toast({
              title: "Stack Created",
              description: `Stack "${newStackName}" created successfully.`
          });
          setIsCreateOpen(false);
          setNewStackName("");
          // Redirect to editor
          router.push(`/stacks/${newStackName}`);
      } catch (error: any) {
          console.error("Failed to create stack", error);
          toast({
              variant: "destructive",
              title: "Error",
              description: error.message || "Failed to create stack."
          });
      } finally {
          setIsCreating(false);
      }
  };

  const openDeleteDialog = (e: React.MouseEvent, stackName: string) => {
      e.preventDefault();
      e.stopPropagation();
      setStackToDelete(stackName);
      setIsDeleteOpen(true);
  };

  const handleDeleteConfirm = async () => {
      if (!stackToDelete) return;

      setIsDeleting(true);
      try {
          await apiClient.deleteCollection(stackToDelete);
          toast({
              title: "Stack Deleted",
              description: `Stack "${stackToDelete}" has been removed.`
          });

          // Optimistic update
          setStacks(prev => prev.filter(s => s.name !== stackToDelete));

          setIsDeleteOpen(false);
          setStackToDelete(null);
      } catch (error: any) {
          console.error("Failed to delete stack", error);
          toast({
              variant: "destructive",
              title: "Error",
              description: error.message || "Failed to delete stack."
          });
      } finally {
          setIsDeleting(false);
      }
  };

  return (
    <div className="space-y-6 p-8 pt-6 h-[calc(100vh-4rem)] overflow-y-auto">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-1">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
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
                        Enter a unique name for your new stack configuration.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="grid gap-2">
                        <Label htmlFor="name">Stack Name</Label>
                        <Input
                            id="name"
                            value={newStackName}
                            onChange={(e) => setNewStackName(e.target.value)}
                            placeholder="my-app-stack"
                            autoFocus
                        />
                    </div>
                </div>
                <DialogFooter>
                    <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                    <Button onClick={handleCreate} disabled={isCreating}>
                        {isCreating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        Create
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
      </div>

      {loading ? (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                  <div key={i} className="h-40 rounded-lg border bg-muted/20 animate-pulse" />
              ))}
          </div>
      ) : stacks.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 border-2 border-dashed rounded-lg bg-muted/10 text-muted-foreground">
              <Layers className="h-12 w-12 mb-4 opacity-20" />
              <h3 className="text-lg font-medium">No Stacks Found</h3>
              <p className="text-sm mt-1">Create a new stack to start organizing your services.</p>
              <Button variant="outline" className="mt-4" onClick={() => setIsCreateOpen(true)}>
                  Create Stack
              </Button>
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
                    <div className="flex items-center gap-2">
                        <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity text-destructive hover:text-destructive hover:bg-destructive/10"
                            onClick={(e) => openDeleteDialog(e, stack.name)}
                            title="Delete Stack"
                        >
                            <Trash2 className="h-3 w-3" />
                        </Button>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center gap-3 mb-4">
                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                            <Layers className="h-6 w-6 text-primary" />
                        </div>
                        <div className="min-w-0 flex-1">
                            <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                            <div className="text-xs text-muted-foreground font-mono truncate">
                                {stack.services?.length || 0} Services
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            {/* Simple heuristic: if it has services, assume configured */}
                            {(stack.services?.length || 0) > 0 ? (
                                <>
                                    <span className="relative flex h-2 w-2">
                                        <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                                    </span>
                                    Configured
                                </>
                            ) : (
                                <>
                                    <span className="relative flex h-2 w-2">
                                        <span className="relative inline-flex rounded-full h-2 w-2 bg-yellow-500"></span>
                                    </span>
                                    Empty
                                </>
                            )}
                        </div>
                        <div>
                            v{stack.version || "1.0.0"}
                        </div>
                    </div>
                </CardContent>
                </Card>
            </Link>
            ))}
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteOpen} onOpenChange={setIsDeleteOpen}>
          <DialogContent>
              <DialogHeader>
                  <DialogTitle className="flex items-center gap-2 text-destructive">
                      <AlertTriangle className="h-5 w-5" />
                      Delete Stack?
                  </DialogTitle>
                  <DialogDescription>
                      Are you sure you want to delete the stack <strong>{stackToDelete}</strong>?
                      This action cannot be undone and will remove all service configurations in this stack.
                  </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                  <Button variant="outline" onClick={() => setIsDeleteOpen(false)}>Cancel</Button>
                  <Button variant="destructive" onClick={handleDeleteConfirm} disabled={isDeleting}>
                      {isDeleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                      Delete Stack
                  </Button>
              </DialogFooter>
          </DialogContent>
      </Dialog>
    </div>
  );
}
