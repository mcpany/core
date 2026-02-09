/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";
import { StackList } from "@/components/stacks/stack-list";
import { Button } from "@/components/ui/button";
import { Plus, Loader2 } from "lucide-react";
import Link from "next/link";
import { useToast } from "@/hooks/use-toast";

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
      // Ensure we handle array response
      if (Array.isArray(res)) {
          setStacks(res);
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

  const handleDelete = async (name: string) => {
      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;
      try {
          await apiClient.deleteCollection(name);
          toast({ title: "Stack Deleted", description: `Stack ${name} removed.` });
          fetchStacks();
      } catch (e) {
          toast({ variant: "destructive", title: "Error", description: "Failed to delete stack." });
      }
  };

  const handleDeploy = async (name: string) => {
      try {
          await apiClient.applyCollection(name);
          toast({ title: "Stack Deployed", description: `Services defined in ${name} have been applied.` });
      } catch (e) {
          toast({ variant: "destructive", title: "Error", description: "Failed to deploy stack." });
      }
  };

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <Button asChild>
            <Link href="/stacks/new">
                <Plus className="mr-2 h-4 w-4" /> Create Stack
            </Link>
        </Button>
      </div>

      {loading ? (
          <div className="flex items-center justify-center py-20">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      ) : (
          <StackList stacks={stacks} onDelete={handleDelete} onDeploy={handleDeploy} />
      )}
    </div>
  );
}
