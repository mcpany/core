/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import { stackManager } from "@/lib/stack-manager";
import { Loader2 } from "lucide-react";
import jsyaml from "js-yaml";

interface CreateStackDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onComplete: () => void;
}

/**
 * CreateStackDialog component.
 * A dialog for creating a new stack by uploading or pasting configuration.
 *
 * @param props - The component props.
 * @param props.open - Whether the dialog is open.
 * @param props.onOpenChange - Callback to change open state.
 * @param props.onComplete - Callback when stack creation is complete.
 * @returns The rendered dialog component.
 */
export function CreateStackDialog({
  open,
  onOpenChange,
  onComplete,
}: CreateStackDialogProps) {
  const [name, setName] = useState("");
  const [content, setContent] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSave = async () => {
    if (!name) {
      toast.error("Stack name is required");
      return;
    }
    if (!content) {
      toast.error("Stack configuration is required");
      return;
    }

    setIsLoading(true);
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let config: any;
      try {
        config = jsyaml.load(content);
      } catch {
        try {
          config = JSON.parse(content);
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        } catch (e2) {
          throw new Error("Invalid format: Must be valid YAML or JSON");
        }
      }

      await stackManager.saveStack(name, config);
      toast.success(`Stack "${name}" created successfully`);
      onComplete();
      onOpenChange(false);
      setName("");
      setContent("");
    } catch (e: unknown) {
      console.error(e);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      toast.error((e as any).message || "Failed to create stack");
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      const text = e.target?.result as string;
      setContent(text);
      // Try to guess name from filename
      if (!name) {
          const fileName = file.name.split('.')[0];
          setName(fileName);
      }
    };
    reader.readAsText(file);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Create New Stack</DialogTitle>
          <DialogDescription>
            Define a new stack by importing a configuration file (YAML/JSON).
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="name">Stack Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="my-stack"
            />
            <p className="text-xs text-muted-foreground">
              Services will be tagged with <code>stack:{name || "..."}</code>
            </p>
          </div>

          <div className="grid gap-2">
            <Label>Configuration</Label>
            <div className="flex items-center gap-2 mb-2">
                <Input
                    type="file"
                    accept=".yaml,.yml,.json"
                    onChange={handleFileUpload}
                    className="cursor-pointer"
                />
            </div>
            <Textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Paste YAML or JSON configuration here..."
              className="font-mono text-xs min-h-[200px]"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isLoading}>
            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Create Stack
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
