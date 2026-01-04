/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, SecretDefinition } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Plus, Trash2, Key } from "lucide-react";
import { format } from "date-fns";

export function SecretManager() {
  const [secrets, setSecrets] = useState<SecretDefinition[]>([]);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newSecret, setNewSecret] = useState<Partial<SecretDefinition>>({});

  useEffect(() => {
    fetchSecrets();
  }, []);

  const fetchSecrets = async () => {
    try {
      const res = await apiClient.listSecrets();
      setSecrets(res.secrets || []);
    } catch (e) {
      console.error("Failed to fetch secrets", e);
    }
  };

  const handleSave = async () => {
    try {
      await apiClient.saveSecret({
          ...newSecret,
          id: newSecret.id || `sec-${Date.now()}`,
          createdAt: new Date().toISOString()
      } as SecretDefinition);
      setIsDialogOpen(false);
      fetchSecrets();
      setNewSecret({});
    } catch (e) {
      console.error("Failed to save secret", e);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await apiClient.deleteSecret(id);
      fetchSecrets();
    } catch (e) {
      console.error("Failed to delete secret", e);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Secrets</h2>
        <Button onClick={() => setIsDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" /> Add Secret
        </Button>
      </div>

      <Card className="backdrop-blur-sm bg-background/50 overflow-hidden">
        <CardHeader>
          <CardTitle>Environment Secrets</CardTitle>
          <CardDescription>Manage sensitive configuration values securely.</CardDescription>
        </CardHeader>
        <CardContent className="p-0 sm:p-6">
          <div className="overflow-x-auto">
             <Table>
                <TableHeader>
                <TableRow>
                    <TableHead className="w-[150px]">Name</TableHead>
                    <TableHead>Key</TableHead>
                    <TableHead className="hidden sm:table-cell">Provider</TableHead>
                    <TableHead className="hidden md:table-cell">Last Used</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                </TableRow>
                </TableHeader>
                <TableBody>
                {secrets.map((secret) => (
                    <TableRow key={secret.id}>
                    <TableCell className="font-medium">
                        <div className="flex items-center">
                           <Key className="h-4 w-4 mr-2 text-muted-foreground shrink-0" />
                           <span className="truncate max-w-[120px] sm:max-w-none">{secret.name}</span>
                        </div>
                        {/* Mobile-only sub-info */}
                        <div className="sm:hidden text-[10px] text-muted-foreground mt-1">
                             {secret.provider || "Custom"} • {secret.lastUsed ? format(new Date(secret.lastUsed), "MMM d") : "Never"}
                        </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs">
                        <span className="truncate max-w-[100px] inline-block sm:max-w-none" title={secret.key}>
                            {secret.key}
                        </span>
                    </TableCell>
                    <TableCell className="hidden sm:table-cell">{secret.provider || "Custom"}</TableCell>
                    <TableCell className="text-muted-foreground text-xs hidden md:table-cell">
                        {secret.lastUsed ? format(new Date(secret.lastUsed), "MMM d, yyyy HH:mm") : "Never"}
                    </TableCell>
                    <TableCell className="text-right">
                        <Button variant="ghost" size="icon" onClick={() => handleDelete(secret.id)}>
                        <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                    </TableCell>
                    </TableRow>
                ))}
                </TableBody>
             </Table>
          </div>
        </CardContent>
      </Card>

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Secret</DialogTitle>
            <DialogDescription>Create a new secret value.</DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-1 sm:grid-cols-4 items-center gap-2 sm:gap-4">
              <Label htmlFor="name" className="text-left sm:text-right">Name</Label>
              <Input
                id="name"
                value={newSecret.name || ""}
                onChange={(e) => setNewSecret({ ...newSecret, name: e.target.value })}
                className="col-span-1 sm:col-span-3"
              />
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-4 items-center gap-2 sm:gap-4">
              <Label htmlFor="key" className="text-left sm:text-right">Key</Label>
              <Input
                id="key"
                value={newSecret.key || ""}
                onChange={(e) => setNewSecret({ ...newSecret, key: e.target.value })}
                placeholder="OPENAI_API_KEY"
                className="col-span-1 sm:col-span-3 font-mono"
              />
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-4 items-center gap-2 sm:gap-4">
              <Label htmlFor="value" className="text-left sm:text-right">Value</Label>
              <Input
                id="value"
                type="password"
                value={newSecret.value || ""}
                onChange={(e) => setNewSecret({ ...newSecret, value: e.target.value })}
                className="col-span-1 sm:col-span-3 font-mono"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleSave}>Save Secret</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
