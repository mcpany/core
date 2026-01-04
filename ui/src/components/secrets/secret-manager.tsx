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
import { Plus, Trash2, Key, Eye, EyeOff } from "lucide-react";
import { format } from "date-fns";

export function SecretManager() {
  const [secrets, setSecrets] = useState<SecretDefinition[]>([]);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newSecret, setNewSecret] = useState<Partial<SecretDefinition>>({});
  const [showValues, setShowValues] = useState<Record<string, boolean>>({});

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

  const toggleVisibility = (id: string) => {
      setShowValues(prev => ({ ...prev, [id]: !prev[id] }));
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Secrets</h2>
        <Button onClick={() => setIsDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" /> Add Secret
        </Button>
      </div>

      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Environment Secrets</CardTitle>
          <CardDescription>Manage sensitive configuration values securely.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>Provider</TableHead>
                <TableHead>Last Used</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {secrets.map((secret) => (
                <TableRow key={secret.id}>
                  <TableCell className="font-medium flex items-center">
                    <Key className="h-4 w-4 mr-2 text-muted-foreground" />
                    {secret.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{secret.key}</TableCell>
                  <TableCell>{secret.provider || "Custom"}</TableCell>
                  <TableCell className="text-muted-foreground text-xs">
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
        </CardContent>
      </Card>

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Secret</DialogTitle>
            <DialogDescription>Create a new secret value.</DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="name" className="text-right">Name</Label>
              <Input
                id="name"
                value={newSecret.name || ""}
                onChange={(e) => setNewSecret({ ...newSecret, name: e.target.value })}
                className="col-span-3"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="key" className="text-right">Key</Label>
              <Input
                id="key"
                value={newSecret.key || ""}
                onChange={(e) => setNewSecret({ ...newSecret, key: e.target.value })}
                placeholder="OPENAI_API_KEY"
                className="col-span-3 font-mono"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="value" className="text-right">Value</Label>
              <Input
                id="value"
                type="password"
                value={newSecret.value || ""}
                onChange={(e) => setNewSecret({ ...newSecret, value: e.target.value })}
                className="col-span-3 font-mono"
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
