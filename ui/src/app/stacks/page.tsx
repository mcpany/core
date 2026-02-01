/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Layers, Cuboid, Plus, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
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
import { apiClient, Collection } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newStackName, setNewStackName] = useState("");
  const [creating, setCreating] = useState(false);
  const router = useRouter();
  const { toast } = useToast();

  const fetchStacks = async () => {
    setLoading(true);
    try {
      const data = await apiClient.listCollections();
      setStacks(data);
    } catch (error) {
      console.error("Failed to fetch stacks", error);
      toast({
          title: "Error",
          description: "Failed to load stacks.",
          variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const handleCreateStack = async () => {
    if (!newStackName.trim()) return;
    setCreating(true);
    try {
      // Initialize with basic structure
      const newStack: Partial<Collection> & { name: string } = {
          name: newStackName,
          description: "Created via UI",
          version: "1.0.0",
          services: [],
          skills: []
      };
      await apiClient.saveCollection(newStack);
      toast({
          title: "Stack Created",
          description: `Stack ${newStackName} has been created.`,
      });
      setIsDialogOpen(false);
      router.push(`/stacks/${newStackName}`);
    } catch (error) {
        console.error("Failed to create stack", error);
        toast({
            title: "Error",
            description: "Failed to create stack.",
            variant: "destructive",
        });
    } finally {
        setCreating(false);
    }
  };

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-1">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
            <DialogTrigger asChild>
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> Create Stack
                </Button>
            </DialogTrigger>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Create New Stack</DialogTitle>
                    <DialogDescription>
                        Enter a name for your new stack configuration.
                    </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="name" className="text-right">
                            Name
                        </Label>
                        <Input
                            id="name"
                            value={newStackName}
                            onChange={(e) => setNewStackName(e.target.value)}
                            className="col-span-3"
                            placeholder="my-stack"
                        />
                    </div>
                </div>
                <DialogFooter>
                    <Button onClick={handleCreateStack} disabled={creating || !newStackName.trim()}>
                        {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        Create
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
      </div>

      {loading ? (
          <div className="flex items-center justify-center h-64">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.length === 0 && (
                <div className="col-span-full text-center p-12 border rounded-lg border-dashed text-muted-foreground bg-muted/10">
                    No stacks found. Create one to get started.
                </div>
            )}
            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
                <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full">
                <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                        Stack
                    </CardTitle>
                    <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                </CardHeader>
                <CardContent>
                    <div className="flex items-center gap-3 mb-4">
                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                            <Layers className="h-6 w-6 text-primary" />
                        </div>
                        <div>
                            <div className="text-2xl font-bold tracking-tight truncate max-w-[200px]" title={stack.name}>{stack.name}</div>
                            {stack.version && <div className="text-xs text-muted-foreground font-mono">v{stack.version}</div>}
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            {/* We assume online for now as stacks are config files, but maybe check if any service running? */}
                            <Badge variant="outline" className="font-normal text-[10px] h-5">
                                Configured
                            </Badge>
                        </div>
                        <div>
                            {stack.services?.length || 0} Services
                        </div>
                    </div>
                </CardContent>
                </Card>
            </Link>
            ))}
        </div>
      )}
    </div>
  );
}
