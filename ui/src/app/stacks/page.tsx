/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { StackList, Stack } from "@/components/stacks/stack-list";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchStacks = async () => {
    setLoading(true);
    try {
      const res = await apiClient.listCollections();
      // res is Collection[]
      // Map to Stack interface
      const mapped: Stack[] = res.map((c: any) => ({
          name: c.name,
          description: c.description,
          author: c.author,
          version: c.version,
          services: c.services || []
      }));
      setStacks(mapped);
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
      if (!confirm(`Are you sure you want to delete stack "${name}"? This will remove the stack definition but may keep services if not properly cleaned up (implementation pending).`)) return;

      try {
          await apiClient.deleteCollection(name);
          toast({ title: "Stack Deleted", description: `Stack ${name} removed.` });
          fetchStacks();
      } catch (e) {
          console.error(e);
          toast({ variant: "destructive", title: "Error", description: "Failed to delete stack." });
      }
  };

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div>
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground mt-2">Manage your MCP Any configuration stacks.</p>
        </div>
        <Button asChild>
            <Link href="/stacks/new">
                <Plus className="mr-2 h-4 w-4" /> Add Stack
            </Link>
        </Button>
      </div>

      <StackList stacks={stacks} isLoading={loading} onDelete={handleDelete} />
    </div>
  );
}
