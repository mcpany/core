/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { stackManager, Stack } from "@/lib/stack-manager";
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Layers, Cuboid, Trash2, ExternalLink } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";

export function StackList() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchStacks = async () => {
    setIsLoading(true);
    try {
      const data = await stackManager.listStacks();
      setStacks(data);
    } catch (e) {
      console.error(e);
      toast.error("Failed to load stacks");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const handleDelete = async (name: string) => {
    if (!confirm(`Are you sure you want to delete stack "${name}"? This will delete all services in it.`)) return;
    try {
      await stackManager.deleteStack(name);
      toast.success(`Stack "${name}" deleted`);
      fetchStacks();
    } catch (e) {
      console.error(e);
      toast.error("Failed to delete stack");
    }
  };

  if (isLoading) {
    return (
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-40 rounded-lg bg-muted/20 animate-pulse" />
        ))}
      </div>
    );
  }

  if (stacks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-lg bg-muted/10 text-muted-foreground">
        <Layers className="h-10 w-10 mb-4 opacity-50" />
        <h3 className="text-lg font-medium">No Stacks Found</h3>
        <p className="text-sm mt-1">Create a stack to group your services.</p>
      </div>
    );
  }

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      {stacks.map((stack) => (
        <Card key={stack.name} className="flex flex-col hover:shadow-md transition-all group border-transparent shadow-sm bg-card hover:bg-muted/50">
          <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
               <Cuboid className="h-4 w-4" /> Stack
            </CardTitle>
            <Badge
                variant={
                    stack.status === "Active" ? "default" :
                    stack.status === "Error" ? "destructive" : "secondary"
                }
                className={stack.status === "Active" ? "bg-green-500 hover:bg-green-600" : ""}
            >
                {stack.status}
            </Badge>
          </CardHeader>
          <CardContent className="flex-1">
             <Link href={`/stacks/${stack.name}`} className="block">
                <div className="flex items-center gap-3 mb-4">
                    <div className="p-3 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                        <Layers className="h-6 w-6 text-primary" />
                    </div>
                    <div>
                        <div className="text-xl font-bold tracking-tight">{stack.name}</div>
                        <div className="text-xs text-muted-foreground font-mono">{stack.services.length} Services</div>
                    </div>
                </div>
             </Link>
          </CardContent>
          <CardFooter className="pt-2 border-t flex justify-between">
              <Link href={`/stacks/${stack.name}`}>
                  <Button variant="ghost" size="sm" className="h-8">
                      <ExternalLink className="mr-2 h-3 w-3" /> Manage
                  </Button>
              </Link>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                onClick={() => handleDelete(stack.name)}
              >
                  <Trash2 className="h-3 w-3" />
              </Button>
          </CardFooter>
        </Card>
      ))}
    </div>
  );
}
