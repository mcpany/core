/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useCallback } from "react";
import { apiClient } from "@/lib/client";
import { StackList } from "@/components/stacks/stack-list";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import Link from "next/link";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchStacks = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiClient.listCollections();
      setStacks(res || []);
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
  }, [toast]);

  useEffect(() => {
    fetchStacks();
  }, [fetchStacks]);

  const handleDelete = async (name: string) => {
      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;
      try {
          await apiClient.deleteCollection(name);
          toast({ title: "Stack Deleted", description: `${name} has been removed.` });
          fetchStacks();
      } catch (e) {
          console.error("Failed to delete stack", e);
          toast({ variant: "destructive", title: "Error", description: "Failed to delete stack." });
      }
  };

  const handleDeploy = async (name: string) => {
      try {
          await apiClient.deployStack(name);
          toast({ title: "Stack Deployed", description: `${name} has been applied.` });
      } catch (e) {
          console.error("Failed to deploy stack", e);
          toast({ variant: "destructive", title: "Error", description: "Failed to deploy stack." });
      }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div>
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground mt-2">Manage and deploy collections of services.</p>
        </div>
        <div className="flex items-center gap-2">
            <Button asChild>
                <Link href="/stacks/new">
                    <Plus className="mr-2 h-4 w-4" /> Add Stack
                </Link>
            </Button>
        </div>
      </div>

      <StackList
        stacks={stacks}
        isLoading={loading}
        onDelete={handleDelete}
        onDeploy={handleDeploy}
      />
    </div>
  );
}
