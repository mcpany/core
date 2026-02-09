/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Plus, Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";

interface CreateStackDialogProps {
    onSuccess?: () => void;
}

/**
 * CreateStackDialog component.
 * Allows users to create a new service collection (stack).
 * @param props - The component props.
 * @returns The rendered component.
 */
export function CreateStackDialog({ onSuccess }: CreateStackDialogProps) {
  const [open, setOpen] = useState(false);
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);

  // Form State
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  const handleSave = async () => {
    if (!name) {
        toast({
            title: "Validation Error",
            description: "Name is required.",
            variant: "destructive"
        });
        return;
    }

    setLoading(true);
    try {
        await apiClient.createCollection({
            name,
            description,
            services: [] // Start empty
        });

        toast({
            title: "Stack Created",
            description: "New stack has been successfully created."
        });
        setOpen(false);
        setName("");
        setDescription("");
        if (onSuccess) onSuccess();
    } catch (error) {
        console.error(error);
        toast({
            title: "Error",
            description: "Failed to create stack.",
            variant: "destructive"
        });
    } finally {
        setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
            <Plus className="mr-2 h-4 w-4" /> Create Stack
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Create New Stack</DialogTitle>
          <DialogDescription>
            Create a collection of services to manage together.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">
              Name
            </Label>
            <Input
                id="name"
                placeholder="e.g. dev-tools"
                className="col-span-3"
                value={name}
                onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-4 items-start gap-4">
            <Label htmlFor="description" className="text-right pt-2">
              Description
            </Label>
            <Textarea
                id="description"
                placeholder="Optional description..."
                className="col-span-3"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)} disabled={loading}>Cancel</Button>
            <Button onClick={handleSave} disabled={loading}>
                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create Stack
            </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
