/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, RefreshCw, Trash2, MoreHorizontal, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface Stack {
    name: string;
    services: any[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newStackName, setNewStackName] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [stackToDelete, setStackToDelete] = useState<string | null>(null);
  const { toast } = useToast();

  const fetchStacks = async () => {
      setLoading(true);
      try {
          const res = await apiClient.listCollections();
          setStacks(res || []);
      } catch (error) {
          console.error("Failed to load stacks", error);
          toast({
              title: "Error",
              description: "Failed to load stacks.",
              variant: "destructive"
          });
      } finally {
          setLoading(false);
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
              description: `Stack "${newStackName}" created successfully.`
          });
          setNewStackName("");
          setIsCreateOpen(false);
          fetchStacks();
      } catch (error) {
          console.error("Failed to create stack", error);
          toast({
              title: "Error",
              description: "Failed to create stack.",
              variant: "destructive"
          });
      } finally {
          setIsCreating(false);
      }
  };

  const handleDelete = async () => {
      if (!stackToDelete) return;
      try {
          await apiClient.deleteCollection(stackToDelete);
          toast({
              title: "Stack Deleted",
              description: `Stack "${stackToDelete}" has been deleted.`
          });
          fetchStacks();
      } catch (error) {
          console.error("Failed to delete stack", error);
          toast({
              title: "Error",
              description: "Failed to delete stack.",
              variant: "destructive"
          });
      } finally {
          setStackToDelete(null);
      }
  };

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={fetchStacks} disabled={loading}>
                <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                Refresh
            </Button>
            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                <DialogTrigger asChild>
                    <Button>
                        <Plus className="mr-2 h-4 w-4" />
                        Add Stack
                    </Button>
                </DialogTrigger>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Create New Stack</DialogTitle>
                        <DialogDescription>
                            Enter a name for the new stack configuration.
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
                        <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                        <Button onClick={handleCreate} disabled={!newStackName.trim() || isCreating}>
                            {isCreating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Create
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
      </div>

      {loading && stacks.length === 0 ? (
          <div className="flex items-center justify-center h-64">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
                <Card key={stack.name} className="hover:shadow-md transition-all group border-transparent shadow-sm bg-card hover:bg-muted/50 flex flex-col">
                    <Link href={`/stacks/${stack.name}`} className="flex-1 cursor-pointer">
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
                                    <div className="text-2xl font-bold tracking-tight truncate max-w-[150px]" title={stack.name}>
                                        {stack.name}
                                    </div>
                                </div>
                            </div>
                        </CardContent>
                    </Link>
                    <CardFooter className="pt-2 border-t bg-muted/20 flex justify-between items-center">
                        <div className="text-xs text-muted-foreground">
                             {stack.services?.length || 0} Services
                        </div>
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                                    <MoreHorizontal className="h-4 w-4" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                                <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setStackToDelete(stack.name)}>
                                    <Trash2 className="mr-2 h-4 w-4" /> Delete
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </CardFooter>
                </Card>
            ))}
            {stacks.length === 0 && (
                <div className="col-span-full flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-lg text-muted-foreground bg-muted/10">
                    <Layers className="h-10 w-10 mb-4 opacity-50" />
                    <h3 className="text-lg font-medium">No stacks found</h3>
                    <p className="text-sm mb-4">Create your first stack to get started.</p>
                    <Button onClick={() => setIsCreateOpen(true)} variant="outline">
                        Create Stack
                    </Button>
                </div>
            )}
        </div>
      )}

      <AlertDialog open={!!stackToDelete} onOpenChange={(open) => !open && setStackToDelete(null)}>
        <AlertDialogContent>
            <AlertDialogHeader>
                <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                <AlertDialogDescription>
                    This action cannot be undone. This will permanently delete the stack
                    <strong> {stackToDelete}</strong> and remove all associated service configurations.
                </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction onClick={handleDelete} className="bg-destructive hover:bg-destructive/90 text-destructive-foreground">
                    Delete
                </AlertDialogAction>
            </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
