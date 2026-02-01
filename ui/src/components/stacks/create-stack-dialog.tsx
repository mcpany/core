/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { apiClient } from "@/lib/client";
import { Loader2 } from "lucide-react";

interface CreateStackDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * Dialog for creating a new stack.
 * @param props - Component props.
 * @returns The rendered dialog.
 */
export function CreateStackDialog({ open, onOpenChange }: CreateStackDialogProps) {
  const [name, setName] = useState("");
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setLoading(true);
    try {
        await apiClient.saveCollection({
            name: name,
            services: []
        });
        toast.success("Stack created successfully");
        onOpenChange(false);
        router.push(`/stacks/${name}`);
    } catch (error) {
        console.error(error);
        toast.error("Failed to create stack");
    } finally {
        setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create New Stack</DialogTitle>
          <DialogDescription>
            Create a new service collection. You can add services to it later.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleCreate}>
            <div className="grid gap-4 py-4">
            <div className="grid gap-2">
                <Label htmlFor="name">Stack Name</Label>
                <Input
                id="name"
                placeholder="e.g. my-stack"
                value={name}
                onChange={(e) => setName(e.target.value)}
                autoFocus
                />
            </div>
            </div>
            <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                Cancel
            </Button>
            <Button type="submit" disabled={loading || !name.trim()}>
                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create
            </Button>
            </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
