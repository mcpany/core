/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Trash2, Loader2, AlertCircle } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";
import { useToast } from "@/hooks/use-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<ServiceCollection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { toast } = useToast();

  const fetchStacks = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiClient.listCollections();
      // Ensure we handle different response structures if needed
      if (Array.isArray(res)) {
          setStacks(res);
      } else if (res && Array.isArray(res.collections)) {
          setStacks(res.collections);
      } else {
          setStacks([]);
      }
    } catch (err: any) {
      console.error("Failed to fetch stacks", err);
      setError("Failed to load stacks. Is the backend running?");
      // Fallback to empty to avoid crash
      setStacks([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const handleDelete = async (e: React.MouseEvent, name: string) => {
      e.preventDefault(); // Prevent navigation
      e.stopPropagation();

      if (!confirm(`Are you sure you want to delete the stack "${name}"?`)) return;

      try {
          await apiClient.deleteCollection(name);
          toast({
              title: "Stack Deleted",
              description: `Stack "${name}" has been removed.`
          });
          fetchStacks();
      } catch (err: any) {
          console.error("Failed to delete stack", err);
          toast({
              variant: "destructive",
              title: "Error",
              description: "Failed to delete stack."
          });
      }
  };

  if (loading && stacks.length === 0) {
      return (
          <div className="flex items-center justify-center h-[calc(100vh-4rem)]">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      );
  }

  return (
    <div className="space-y-6 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col overflow-y-auto">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
          <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <CreateStackDialog onStackCreated={fetchStacks} />
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>
            {error}
          </AlertDescription>
        </Alert>
      )}

      {stacks.length === 0 && !loading && !error && (
        <div className="flex flex-col items-center justify-center h-64 text-center border-2 border-dashed rounded-lg">
             <Layers className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
             <h3 className="text-lg font-medium">No collections found</h3>
             <p className="text-sm text-muted-foreground mb-4">Create your first stack to get started.</p>
        </div>
      )}

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
                       <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity hover:text-destructive hover:bg-destructive/10"
                            onClick={(e) => handleDelete(e, stack.name)}
                       >
                           <Trash2 className="h-3 w-3" />
                       </Button>
                  </div>
               </CardHeader>
               <CardContent className="flex-1 flex flex-col justify-between">
                 <div>
                     <div className="flex items-center gap-3 mb-4">
                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                            <Layers className="h-6 w-6 text-primary" />
                        </div>
                        <div>
                            <div className="text-xl font-bold tracking-tight truncate max-w-[180px]" title={stack.name}>{stack.name}</div>
                            <div className="text-xs text-muted-foreground line-clamp-1">{stack.description || "No description"}</div>
                        </div>
                     </div>
                 </div>

                 <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                    <div className="flex items-center gap-1.5">
                        <Badge variant="outline" className="font-normal text-[10px]">
                            v{stack.version || "0.0.1"}
                        </Badge>
                        <span className="opacity-50">by {stack.author || "Unknown"}</span>
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
    </div>
  );
}
