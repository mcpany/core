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
import { Plus, Loader2 } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";

interface CreateStackDialogProps {
    onStackCreated: () => void;
}

/**
 * CreateStackDialog component.
 * Allows creating a new empty stack (service collection).
 * @param props - The component props.
 * @returns The rendered component.
 */
export function CreateStackDialog({ onStackCreated }: CreateStackDialogProps) {
  const [open, setOpen] = useState(false);
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [name, setName] = useState("");

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
        await apiClient.saveCollection({
            name,
            services: [] // Start empty
        });

        toast({
            title: "Stack Created",
            description: `Stack "${name}" has been successfully created.`
        });
        setOpen(false);
        setName("");
        onStackCreated();
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
            <Plus className="mr-2 h-4 w-4" /> New Stack
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create New Stack</DialogTitle>
          <DialogDescription>
            Create a new collection of services.
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
        </div>
        <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)} disabled={loading}>Cancel</Button>
            <Button onClick={handleSave} disabled={loading}>
                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create
            </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
