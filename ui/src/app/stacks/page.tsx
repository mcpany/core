/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { StackList, Stack } from "@/components/stacks/stack-list";
import { useToast } from "@/hooks/use-toast";

export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchStacks = async () => {
    try {
      setLoading(true);
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
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const handleDelete = async (name: string) => {
      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;
      try {
          await apiClient.deleteCollection(name);
          toast({ title: "Stack Deleted", description: `Stack ${name} has been removed.` });
          fetchStacks();
      } catch (e) {
          toast({ variant: "destructive", title: "Error", description: "Failed to delete stack." });
      }
  };

  const handleExport = (stack: Stack) => {
      const blob = new Blob([JSON.stringify(stack, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `${stack.name}-stack.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div>
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground mt-2">Manage your MCP Any configuration stacks.</p>
        </div>
        <Link href="/stacks/new">
            <Button>
                <Plus className="mr-2 h-4 w-4" /> Add Stack
            </Button>
        </Link>
      </div>

      <StackList
        stacks={stacks}
        isLoading={loading}
        onDelete={handleDelete}
        onExport={handleExport}
      />
    </div>
  );
}
