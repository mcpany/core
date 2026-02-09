/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiClient } from "@/lib/client";
import { StackList } from "@/components/stacks/stack-list";
import { StackEditor } from "@/components/stacks/stack-editor";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
    DialogFooter,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useToast } from "@/hooks/use-toast";
import { Collection } from "@proto/config/v1/collection";

export default function StacksPage() {
  const [stacks, setStacks] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedStackId, setSelectedStackId] = useState<string | null>(null);
  const [isEditorOpen, setIsEditorOpen] = useState(false);

  // Creation State
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newStackName, setNewStackName] = useState("");
  const [isCreating, setIsCreating] = useState(false);

  // Deletion State
  const [stackToDelete, setStackToDelete] = useState<string | null>(null);

  const { toast } = useToast();

  useEffect(() => {
    fetchStacks();
  }, []);

  const fetchStacks = async () => {
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
  };

  const handleEdit = (stack: Collection) => {
      setSelectedStackId(stack.name);
      setIsEditorOpen(true);
  };

  const handleCreateOpen = () => {
      setNewStackName("");
      setIsCreateOpen(true);
  };

  const handleCreateSubmit = async () => {
      if (!newStackName.trim()) return;
      if (!/^[a-zA-Z0-9_-]+$/.test(newStackName)) {
          toast({ variant: "destructive", title: "Invalid Name", description: "Use only letters, numbers, dashes, and underscores." });
          return;
      }

      setIsCreating(true);
      try {
          const newStack = {
              name: newStackName,
              description: "New Stack",
              version: "1.0",
              services: []
          };
          await apiClient.saveCollection(newStack);
          await fetchStacks();
          setIsCreateOpen(false);
          // Open editor immediately
          setSelectedStackId(newStackName);
          setIsEditorOpen(true);
      } catch (e) {
          toast({ variant: "destructive", title: "Error", description: "Failed to create stack." });
      } finally {
          setIsCreating(false);
      }
  };

  const handleDeleteConfirm = async () => {
      if (!stackToDelete) return;
      try {
          await apiClient.deleteCollection(stackToDelete);
          fetchStacks();
          toast({ title: "Stack Deleted", description: `Stack ${stackToDelete} removed.` });
      } catch (e) {
          toast({ variant: "destructive", title: "Error", description: "Failed to delete stack." });
      } finally {
          setStackToDelete(null);
      }
  };

  const handleApply = async (name: string) => {
      try {
          await apiClient.applyCollection(name);
          toast({ title: "Stack Deployed", description: `Services from ${name} have been registered.` });
      } catch (e) {
          toast({ variant: "destructive", title: "Deployment Failed", description: String(e) });
      }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div>
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground mt-2">
                Manage Service Collections (Stacks) to deploy multiple services at once.
            </p>
        </div>
        <div className="flex items-center gap-2">
            <Button onClick={handleCreateOpen}>
                <Plus className="mr-2 h-4 w-4" /> New Stack
            </Button>
        </div>
      </div>

      <StackList
        stacks={stacks}
        isLoading={loading}
        onEdit={handleEdit}
        onDelete={setStackToDelete}
        onApply={handleApply}
      />

      {/* Editor Dialog */}
      <Dialog open={isEditorOpen} onOpenChange={(open) => {
          setIsEditorOpen(open);
          if (!open) {
              fetchStacks(); // Refresh list on close
              setSelectedStackId(null);
          }
      }}>
        <DialogContent className="max-w-7xl h-[90vh] flex flex-col p-0 gap-0 overflow-hidden">
             {selectedStackId && <StackEditor stackId={selectedStackId} />}
        </DialogContent>
      </Dialog>

      {/* Creation Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogContent>
              <DialogHeader>
                  <DialogTitle>Create New Stack</DialogTitle>
                  <DialogDescription>
                      Enter a unique name for your new stack.
                  </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                      <Label htmlFor="name">Name</Label>
                      <Input
                          id="name"
                          value={newStackName}
                          onChange={(e) => setNewStackName(e.target.value)}
                          placeholder="my-stack"
                          onKeyDown={(e) => e.key === "Enter" && handleCreateSubmit()}
                      />
                  </div>
              </div>
              <DialogFooter>
                  <Button variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
                  <Button onClick={handleCreateSubmit} disabled={isCreating || !newStackName}>
                      {isCreating ? "Creating..." : "Create"}
                  </Button>
              </DialogFooter>
          </DialogContent>
      </Dialog>

      {/* Deletion Alert */}
      <AlertDialog open={!!stackToDelete} onOpenChange={(open) => !open && setStackToDelete(null)}>
        <AlertDialogContent>
            <AlertDialogHeader>
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
                This will delete the stack definition for <b>{stackToDelete}</b>.
                Running services deployed from this stack will <b>not</b> be removed automatically.
            </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDeleteConfirm} className="bg-destructive hover:bg-destructive/90">
                Delete
            </AlertDialogAction>
            </AlertDialogFooter>
        </AlertDialogContent>
        </AlertDialog>
    </div>
  );
}
