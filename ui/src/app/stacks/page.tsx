/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, Loader2, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [collections, setCollections] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchCollections = async () => {
    setLoading(true);
    try {
      const data = await apiClient.listCollections();
      // data might be array or { collections: [] } depending on API consistency
      const list = Array.isArray(data) ? data : (data.collections || []);
      setCollections(list);
    } catch (error) {
      console.error("Failed to fetch collections", error);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to load stacks.",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCollections();
  }, []);

  const handleDelete = async (e: React.MouseEvent, name: string) => {
    e.preventDefault(); // Prevent link navigation
    e.stopPropagation();

    if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

    try {
        await apiClient.deleteCollection(name);
        toast({
            title: "Stack Deleted",
            description: `Stack ${name} has been removed.`
        });
        fetchCollections();
    } catch (e: any) {
        console.error("Failed to delete stack", e);
         toast({
            variant: "destructive",
            title: "Error",
            description: "Failed to delete stack."
        });
    }
  };

  return (
    <div className="space-y-6 p-8">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
          <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
          <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <CreateStackDialog onSuccess={fetchCollections} />
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {collections.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
               <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full flex flex-col relative">
                 <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                      Stack
                    </CardTitle>
                    <div className="flex items-center gap-2">
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-muted-foreground hover:text-destructive z-10"
                            onClick={(e) => handleDelete(e, stack.name)}
                            title="Delete Stack"
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                        <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                    </div>
                 </CardHeader>
                 <CardContent className="flex-1 flex flex-col">
                   <div className="flex items-center gap-3 mb-4">
                      <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                          <Layers className="h-6 w-6 text-primary" />
                      </div>
                      <div className="min-w-0">
                          <div className="text-xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                          <div className="text-xs text-muted-foreground font-mono truncate">{stack.version || "latest"}</div>
                      </div>
                   </div>

                   {stack.description && (
                       <p className="text-sm text-muted-foreground mb-4 line-clamp-2 flex-1">
                           {stack.description}
                       </p>
                   )}

                   <div className="flex items-center justify-between text-xs text-muted-foreground mt-auto pt-4 border-t w-full">
                      <div className="flex items-center gap-1.5">
                          <span className="relative flex h-2 w-2">
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                          </span>
                          Active
                      </div>
                      <div>
                          {(stack.services?.length || 0)} Services
                      </div>
                   </div>
                 </CardContent>
               </Card>
            </Link>
          ))}
          {collections.length === 0 && (
            <div className="col-span-full flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-muted/20 text-muted-foreground">
                <Layers className="h-10 w-10 mb-4 opacity-50" />
                <p>No stacks found. Create one to get started.</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
