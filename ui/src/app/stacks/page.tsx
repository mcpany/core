/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Layers, Cuboid, Plus, Loader2, RefreshCw } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
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
import { apiClient } from "@/lib/client";
import { toast } from "sonner";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newStackName, setNewStackName] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const router = useRouter();

  const fetchStacks = async () => {
    setLoading(true);
    try {
      const data = await apiClient.listCollections();
      // data is array of collections
      setStacks(data || []);
    } catch (error) {
      console.error("Failed to fetch stacks", error);
      toast.error("Failed to load stacks");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newStackName.trim()) return;

    setIsCreating(true);
    try {
      // name is the ID essentially
      await apiClient.saveCollection({ name: newStackName, services: [] });
      toast.success("Stack created successfully");
      setIsCreateOpen(false);
      setNewStackName("");
      // Redirect to the new stack
      router.push(`/stacks/${newStackName}`);
    } catch (error) {
      console.error("Failed to create stack", error);
      toast.error("Failed to create stack");
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={fetchStacks} disabled={loading}>
                <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} /> Refresh
            </Button>
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
                            Enter a name for your new service collection.
                        </DialogDescription>
                    </DialogHeader>
                    <form onSubmit={handleCreate}>
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
                                    autoFocus
                                />
                            </div>
                        </div>
                        <DialogFooter>
                            <Button type="submit" disabled={!newStackName.trim() || isCreating}>
                                {isCreating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                Create
                            </Button>
                        </DialogFooter>
                    </form>
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
            {stacks.length === 0 && (
                <div className="col-span-full text-center py-12 text-muted-foreground">
                    No stacks found. Create one to get started.
                </div>
            )}
            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
                <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full">
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
                            <div className="text-2xl font-bold tracking-tight truncate max-w-[180px]">{stack.name}</div>
                            {stack.description && (
                                <div className="text-xs text-muted-foreground line-clamp-1">{stack.description}</div>
                            )}
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            {/* Placeholder for status logic if we want list-level status */}
                            <span className="relative flex h-2 w-2">
                                <span className="relative inline-flex rounded-full h-2 w-2 bg-slate-300 dark:bg-slate-700"></span>
                            </span>
                            Details
                        </div>
                        <div>
                            {stack.services ? stack.services.length : 0} Services
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
