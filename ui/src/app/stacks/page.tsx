/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, Edit, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";
import { useToast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<ServiceCollection[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchStacks = async () => {
    setLoading(true);
    try {
      const res = await apiClient.listCollections();
      // API returns { collections: [...] } or just [...]?
      // client.ts says: return res.json();
      // api.go handleCollections returns list or wrapped?
      // api.go: buf = append(buf, '[') ... loop ... append(buf, ']')
      // So it returns an array.
      if (Array.isArray(res)) {
          setStacks(res);
      } else if (res && Array.isArray((res as any).collections)) {
          setStacks((res as any).collections);
      } else {
          setStacks([]);
      }
    } catch (e) {
      console.error("Failed to fetch stacks", e);
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

  const handleDelete = async (name: string, e: React.MouseEvent) => {
      e.preventDefault(); // Prevent link navigation
      e.stopPropagation();
      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

      try {
          await apiClient.deleteCollection(name);
          toast({ title: "Stack Deleted", description: `Stack ${name} has been removed.` });
          fetchStacks();
      } catch (err) {
          console.error(err);
          toast({ variant: "destructive", title: "Error", description: "Failed to delete stack." });
      }
  };

  return (
    <div className="space-y-6 p-8">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <Link href="/stacks/new">
            <Button>
                <Plus className="mr-2 h-4 w-4" />
                Add Stack
            </Button>
        </Link>
      </div>

      {loading && (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map(i => (
                  <div key={i} className="h-40 rounded-xl bg-muted/20 animate-pulse" />
              ))}
          </div>
      )}

      {!loading && stacks.length === 0 && (
          <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-muted/10">
              <Layers className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
              <h3 className="text-lg font-medium">No stacks found</h3>
              <p className="text-sm text-muted-foreground mb-4">Create a new stack to get started.</p>
              <Link href="/stacks/new">
                  <Button variant="outline">Create Stack</Button>
              </Link>
          </div>
      )}

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {stacks.map((stack) => (
          <Link key={stack.name} href={`/stacks/${stack.name}`}>
             <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full flex flex-col">
               <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                  <CardTitle className="text-sm font-medium text-muted-foreground truncate w-full pr-4" title={stack.name}>
                    {stack.name}
                  </CardTitle>
                  <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors shrink-0" />
               </CardHeader>
               <CardContent className="flex-1">
                 <div className="flex items-center gap-3 mb-4">
                    <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                        <Layers className="h-6 w-6 text-primary" />
                    </div>
                    <div className="overflow-hidden">
                        <div className="text-xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                        <div className="text-xs text-muted-foreground font-mono truncate">{stack.version || "latest"}</div>
                    </div>
                 </div>
                 <CardDescription className="line-clamp-2 min-h-[40px]">
                     {stack.description || "No description provided."}
                 </CardDescription>
               </CardContent>
               <CardFooter className="pt-0 border-t bg-muted/20 p-4 flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <Badge variant="secondary" className="text-xs font-normal">
                            {stack.services?.length || 0} Services
                        </Badge>
                    </div>
                    <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-foreground" asChild onClick={(e) => e.stopPropagation()}>
                             <Link href={`/stacks/${stack.name}`}>
                                <Edit className="h-4 w-4" />
                             </Link>
                        </Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10" onClick={(e) => handleDelete(stack.name, e)}>
                            <Trash2 className="h-4 w-4" />
                        </Button>
                    </div>
               </CardFooter>
             </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
