/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Key, Plus, Trash2, Shield, Eye, EyeOff, RefreshCw, Search, Copy } from "lucide-react";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Checkbox } from "@/components/ui/checkbox";

interface SecretDefinition {
    id: string;
    name: string;
    key: string;
    provider: string;
}

/**
 * SecretsManager component.
 * @returns The rendered component.
 */
export function SecretsManager() {
    const [secrets, setSecrets] = useState<SecretDefinition[]>([]);
    const [loading, setLoading] = useState(true);
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const { toast } = useToast();

    // Selected state for bulk actions
    const [selectedSecrets, setSelectedSecrets] = useState<Set<string>>(new Set());
    const [isDeleting, setIsDeleting] = useState(false);

    // New secret state
    const [newSecretName, setNewSecretName] = useState("");
    const [newSecretKey, setNewSecretKey] = useState("");
    const [newSecretValue, setNewSecretValue] = useState("");
    const [newSecretProvider, setNewSecretProvider] = useState("custom");

    useEffect(() => {
        fetchSecrets();
    }, []);

    const fetchSecrets = async () => {
        setLoading(true);
        try {
            const data = await apiClient.listSecrets();
            if (Array.isArray(data)) {
                 setSecrets(data);
            } else if (data && Array.isArray(data.secrets)) {
                setSecrets(data.secrets);
            } else {
                setSecrets([]);
            }
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: "Failed to load secrets", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    };

    const handleSaveSecret = async () => {
        if (!newSecretName || !newSecretKey || !newSecretValue) {
            toast({ title: "Error", description: "Name, key, and value are required.", variant: "destructive" });
            return;
        }

        try {
            await apiClient.createSecret({
                name: newSecretName,
                key: newSecretKey,
                value: newSecretValue,
                provider: newSecretProvider
            });
            toast({ title: "Success", description: "Secret created." });
            setIsAddDialogOpen(false);
            resetForm();
            fetchSecrets();
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: "Failed to create secret", variant: "destructive" });
        }
    };

    const handleDeleteSecret = async (id: string) => {
        try {
            await apiClient.deleteSecret(id);
            toast({ title: "Success", description: "Secret deleted." });

            // Remove from selection if selected
            setSelectedSecrets(prev => {
                const next = new Set(prev);
                next.delete(id);
                return next;
            });

            fetchSecrets();
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: "Failed to delete secret", variant: "destructive" });
        }
    };

    const handleBulkDelete = async () => {
        if (selectedSecrets.size === 0) return;
        setIsDeleting(true);

        try {
            // Call API concurrently for all selected secrets
            const promises = Array.from(selectedSecrets).map(id => apiClient.deleteSecret(id));
            await Promise.all(promises);

            toast({
                title: "Success",
                description: `${selectedSecrets.size} secret(s) deleted.`
            });
            setSelectedSecrets(new Set());
            fetchSecrets();
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: "Failed to delete one or more secrets.", variant: "destructive" });
            // Even if one fails, we should refresh to get the current state
            fetchSecrets();
        } finally {
            setIsDeleting(false);
        }
    };

    const resetForm = () => {
        setNewSecretName("");
        setNewSecretKey("");
        setNewSecretValue("");
        setNewSecretProvider("custom");
    };

    const filteredSecrets = secrets.filter(s =>
        s.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        s.key.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const isAllSelected = filteredSecrets.length > 0 && selectedSecrets.size === filteredSecrets.length;

    const toggleSelectAll = (checked: boolean) => {
        if (checked) {
            setSelectedSecrets(new Set(filteredSecrets.map(s => s.id)));
        } else {
            setSelectedSecrets(new Set());
        }
    };

    const toggleSelect = (id: string, checked: boolean) => {
        setSelectedSecrets(prev => {
            const next = new Set(prev);
            if (checked) next.add(id);
            else next.delete(id);
            return next;
        });
    };

    return (
        <div className="flex flex-col h-full space-y-4 max-w-5xl mx-auto w-full">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Secrets Vault</h2>
                    <p className="text-sm text-muted-foreground">
                        Manage secure credentials for your upstream services.
                    </p>
                </div>
                <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
                    <DialogTrigger asChild>
                        <Button onClick={resetForm}>
                            <Plus className="mr-2 h-4 w-4" /> Add Secret
                        </Button>
                    </DialogTrigger>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>Add New Secret</DialogTitle>
                            <DialogDescription>
                                Securely store an API key or credential.
                            </DialogDescription>
                        </DialogHeader>
                        <div className="grid gap-4 py-4">
                            <div className="grid gap-2">
                                <Label htmlFor="provider">Provider</Label>
                                <Select value={newSecretProvider} onValueChange={setNewSecretProvider}>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select provider" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="custom">Custom</SelectItem>
                                        <SelectItem value="openai">OpenAI</SelectItem>
                                        <SelectItem value="anthropic">Anthropic</SelectItem>
                                        <SelectItem value="aws">AWS</SelectItem>
                                        <SelectItem value="gcp">Google Cloud</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="name">Friendly Name</Label>
                                <Input
                                    id="name"
                                    placeholder="e.g. Production OpenAI Key"
                                    value={newSecretName}
                                    onChange={(e) => setNewSecretName(e.target.value)}
                                />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="key">Key Name (Env Var)</Label>
                                <Input
                                    id="key"
                                    placeholder="e.g. OPENAI_API_KEY"
                                    value={newSecretKey}
                                    onChange={(e) => setNewSecretKey(e.target.value)}
                                />
                            </div>
                            <div className="grid gap-2">
                                <Label htmlFor="value">Secret Value</Label>
                                <Input
                                    id="value"
                                    type="password"
                                    placeholder="sk-..."
                                    value={newSecretValue}
                                    onChange={(e) => setNewSecretValue(e.target.value)}
                                />
                            </div>
                        </div>
                        <DialogFooter>
                            <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>Cancel</Button>
                            <Button onClick={handleSaveSecret}>Save Secret</Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>

            <Card className="flex-1 flex flex-col overflow-hidden bg-background/50 backdrop-blur-sm border-muted/50">
                <CardHeader className="p-4 border-b bg-muted/20 flex flex-row items-center justify-between space-y-0">
                     <div className="relative">
                        <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            placeholder="Search secrets..."
                            className="pl-8 bg-background w-[300px]"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>
                    {selectedSecrets.size > 0 && (
                        <div className="flex items-center gap-4 bg-muted/50 px-4 py-1.5 rounded-md border border-muted/80 animate-in fade-in duration-200">
                            <span className="text-sm font-medium text-muted-foreground">
                                {selectedSecrets.size} selected
                            </span>
                            <Button
                                variant="destructive"
                                size="sm"
                                className="h-7 text-xs"
                                onClick={handleBulkDelete}
                                disabled={isDeleting}
                            >
                                {isDeleting ? <RefreshCw className="mr-2 h-3 w-3 animate-spin" /> : <Trash2 className="mr-2 h-3 w-3" />}
                                Bulk Delete
                            </Button>
                        </div>
                    )}
                </CardHeader>
                <CardContent className="p-0 flex-1 overflow-hidden">
                    <ScrollArea className="h-full">
                        {loading ? (
                            <div className="flex items-center justify-center h-40 text-muted-foreground gap-2">
                                <RefreshCw className="h-4 w-4 animate-spin" /> Loading secrets...
                            </div>
                        ) : filteredSecrets.length === 0 ? (
                            <div className="flex flex-col items-center justify-center h-40 text-muted-foreground gap-2">
                                <Shield className="h-8 w-8 opacity-20" />
                                <p>No secrets found.</p>
                            </div>
                        ) : (
                            <Table>
                                <TableHeader>
                                    <TableRow className="hover:bg-transparent">
                                        <TableHead className="w-[40px] pl-4">
                                            <Checkbox
                                                checked={isAllSelected}
                                                onCheckedChange={(checked) => toggleSelectAll(!!checked)}
                                                aria-label="Select all secrets"
                                            />
                                        </TableHead>
                                        <TableHead>Name</TableHead>
                                        <TableHead>Provider</TableHead>
                                        <TableHead>Key (Env Var)</TableHead>
                                        <TableHead>Value</TableHead>
                                        <TableHead className="text-right pr-4">Actions</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {filteredSecrets.map((secret) => (
                                        <SecretItemRow
                                            key={secret.id}
                                            secret={secret}
                                            isSelected={selectedSecrets.has(secret.id)}
                                            onToggleSelect={(checked) => toggleSelect(secret.id, checked)}
                                            onDelete={handleDeleteSecret}
                                        />
                                    ))}
                                </TableBody>
                            </Table>
                        )}
                    </ScrollArea>
                </CardContent>
            </Card>
        </div>
    );
}

/**
 * SecretItemRow component for Table rendering.
 */
function SecretItemRow({
    secret,
    isSelected,
    onToggleSelect,
    onDelete
}: {
    secret: SecretDefinition;
    isSelected: boolean;
    onToggleSelect: (checked: boolean) => void;
    onDelete: (id: string) => void;
}) {
    const [revealedValue, setRevealedValue] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const handleReveal = async () => {
        if (revealedValue) {
            setRevealedValue(null);
            return;
        }
        setLoading(true);
        try {
            const res = await apiClient.revealSecret(secret.id);
            setRevealedValue(res.value);
        } catch (e) {
            console.error(e);
            toast({ title: "Error", description: "Failed to reveal secret", variant: "destructive" });
        } finally {
            setLoading(false);
        }
    };

    const handleCopy = async () => {
        let value = revealedValue;
        if (!value) {
            setLoading(true);
            try {
                const res = await apiClient.revealSecret(secret.id);
                value = res.value;
                setRevealedValue(value);
            } catch (e) {
                console.error(e);
                toast({ title: "Error", description: "Failed to copy secret", variant: "destructive" });
                setLoading(false);
                return;
            }
            setLoading(false);
        }

        if (value) {
            navigator.clipboard.writeText(value);
            toast({
                title: "Copied",
                description: "Secret value copied to clipboard.",
            });
        }
    };

    return (
        <TableRow data-state={isSelected ? "selected" : undefined} className="group">
            <TableCell className="pl-4">
                <Checkbox
                    checked={isSelected}
                    onCheckedChange={(checked) => onToggleSelect(!!checked)}
                    aria-label={`Select ${secret.name}`}
                />
            </TableCell>
            <TableCell className="font-medium">
                <div className="flex items-center gap-2">
                    <Key className="h-4 w-4 text-primary opacity-70" />
                    {secret.name}
                </div>
            </TableCell>
            <TableCell>
                <Badge variant="outline" className="text-[10px] h-5 font-mono capitalize">
                    {secret.provider}
                </Badge>
            </TableCell>
            <TableCell>
                <span className="text-xs font-mono text-muted-foreground">{secret.key}</span>
            </TableCell>
            <TableCell>
                <div className="flex items-center gap-2 bg-muted/30 rounded-md px-2 py-1 border font-mono text-xs w-[180px] justify-between">
                    <span className="truncate text-muted-foreground">
                        {loading ? <RefreshCw className="h-3 w-3 animate-spin" /> :
                            revealedValue ? revealedValue : "••••••••••••••••••••••••"}
                    </span>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-5 w-5 hover:bg-muted"
                        onClick={handleReveal}
                        disabled={loading}
                        aria-label={revealedValue ? "Hide secret" : "Show secret"}
                    >
                        {revealedValue ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
                    </Button>
                </div>
            </TableCell>
            <TableCell className="text-right pr-4">
                <div className="flex justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleCopy} disabled={loading} aria-label="Copy secret">
                        <Copy className="h-4 w-4 text-muted-foreground" />
                    </Button>
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive/70 hover:text-destructive hover:bg-destructive/10" onClick={() => onDelete(secret.id)} aria-label="Delete secret">
                        <Trash2 className="h-4 w-4" />
                    </Button>
                </div>
            </TableCell>
        </TableRow>
    );
}
