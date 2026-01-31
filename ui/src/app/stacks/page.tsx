/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Layers, Cuboid, Plus, Trash2, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { toast } from "sonner";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const fetchStacks = async () => {
    setLoading(true);
    try {
        const res = await apiClient.listCollections();
        setStacks(res || []);
    } catch (e) {
        console.error(e);
        toast.error("Failed to load stacks");
    } finally {
        setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const handleDelete = async (e: React.MouseEvent, name: string) => {
      e.preventDefault();
      e.stopPropagation(); // Prevent Link navigation
      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

      try {
          await apiClient.deleteCollection(name);
          toast.success("Stack deleted");
          fetchStacks();
      } catch (e) {
          console.error(e);
          toast.error("Failed to delete stack");
      }
  };

  return (
    <div className="space-y-6 p-8">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <Button onClick={() => setIsCreateOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create Stack
        </Button>
      </div>

      {loading ? (
          <div className="flex items-center justify-center h-64">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
                <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50">
                <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                        Stack
                    </CardTitle>
                    <div className="flex items-center gap-2">
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity hover:text-destructive"
                            onClick={(e) => handleDelete(e, stack.name)}
                        >
                            <Trash2 className="h-3 w-3" />
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
                            <div className="text-2xl font-bold tracking-tight">{stack.name}</div>
                            {/* Use ID if available, else name */}
                            <div className="text-xs text-muted-foreground font-mono">{stack.id || stack.name}</div>
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            <span className="relative flex h-2 w-2">
                                {/* Use real status if available later, for now simulate 'Online' if services > 0 */}
                                {stack.services?.length > 0 ? (
                                    <>
                                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                                        <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                                    </>
                                ) : (
                                    <span className="relative inline-flex rounded-full h-2 w-2 bg-slate-300"></span>
                                )}
                            </span>
                            {stack.services?.length > 0 ? "Active" : "Empty"}
                        </div>
                        <div>
                            {stack.services?.length || 0} Services
                        </div>
                    </div>
                </CardContent>
                </Card>
            </Link>
            ))}
            {stacks.length === 0 && (
                <div className="col-span-full flex flex-col items-center justify-center p-12 border border-dashed rounded-lg bg-muted/20 text-muted-foreground">
                    <Layers className="h-12 w-12 mb-4 opacity-50" />
                    <h3 className="text-lg font-medium">No Stacks Found</h3>
                    <p className="text-sm mb-4">Create a new stack to get started.</p>
                    <Button variant="outline" onClick={() => setIsCreateOpen(true)}>Create Stack</Button>
                </div>
            )}
        </div>
      )}

      <CreateStackDialog open={isCreateOpen} onOpenChange={(v) => {
          setIsCreateOpen(v);
          if (!v) fetchStacks(); // Refresh list on close in case one was created
      }} />
    </div>
  );
}
