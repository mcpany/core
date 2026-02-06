/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import {
    Plus,
    Trash2,
    Eye,
    EyeOff,
    Copy,
    Key,
    Shield,
    Search,
    RefreshCw
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useToast } from "@/hooks/use-toast";
import { apiClient, SecretDefinition } from "@/lib/client";
import { Checkbox } from "@/components/ui/checkbox";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
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

/**
 * SecretsManager component.
 * @returns The rendered component.
 */
export function SecretsManager() {
    const [secrets, setSecrets] = useState<SecretDefinition[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState("");
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
    const [selectedSecrets, setSelectedSecrets] = useState<Set<string>>(new Set());
    const [isBulkDeleteDialogOpen, setIsBulkDeleteDialogOpen] = useState(false);
    const { toast } = useToast();

    // Form state
    const [newSecretName, setNewSecretName] = useState("");
    const [newSecretKey, setNewSecretKey] = useState("");
    const [newSecretValue, setNewSecretValue] = useState("");
    const [newSecretProvider, setNewSecretProvider] = useState<string>("custom");

    useEffect(() => {
        loadSecrets();
    }, []);

    const loadSecrets = async () => {
        setLoading(true);
        try {
            const data = await apiClient.listSecrets();
            setSecrets(data);
        } catch (error) {
            console.error("Failed to load secrets", error);
            toast({
                title: "Error",
                description: "Failed to load secrets.",
                variant: "destructive",
            });
        } finally {
            setLoading(false);
        }
    };

    const handleSaveSecret = async () => {
        if (!newSecretName || !newSecretKey || !newSecretValue) {
            toast({
                title: "Validation Error",
                description: "All fields are required.",
                variant: "destructive",
            });
            return;
        }

        try {
            const newSecret: SecretDefinition = {
                id: Math.random().toString(36).substring(7),
                name: newSecretName,
                key: newSecretKey,
                value: newSecretValue,
                provider: newSecretProvider as any,
                createdAt: new Date().toISOString(),
                lastUsed: "Never"
            };

            await apiClient.saveSecret(newSecret);

            toast({
                title: "Success",
                description: "Secret saved successfully.",
            });

            setIsAddDialogOpen(false);
            resetForm();
            loadSecrets();
        } catch (_error) {
            toast({
                title: "Error",
                description: "Failed to save secret.",
                variant: "destructive",
            });
        }
    };

    const handleDeleteSecret = async (id: string) => {
        try {
            await apiClient.deleteSecret(id);
            toast({
                title: "Success",
                description: "Secret deleted successfully.",
            });
            // Remove from selection if it was selected
            if (selectedSecrets.has(id)) {
                const newSelected = new Set(selectedSecrets);
                newSelected.delete(id);
                setSelectedSecrets(newSelected);
            }
            loadSecrets();
        } catch (_error) {
            toast({
                title: "Error",
                description: "Failed to delete secret.",
                variant: "destructive",
            });
        }
    };

    const handleBulkDelete = async () => {
        try {
            const idsToDelete = Array.from(selectedSecrets);
            await Promise.all(idsToDelete.map(id => apiClient.deleteSecret(id)));

            toast({
                title: "Success",
                description: `${idsToDelete.length} secrets deleted successfully.`,
            });
            setSelectedSecrets(new Set());
            setIsBulkDeleteDialogOpen(false);
            loadSecrets();
        } catch (error) {
            console.error("Bulk delete failed", error);
            toast({
                title: "Error",
                description: "Failed to delete some secrets.",
                variant: "destructive",
            });
            setIsBulkDeleteDialogOpen(false);
            loadSecrets();
        }
    };

    const resetForm = () => {
        setNewSecretName("");
        setNewSecretKey("");
        setNewSecretValue("");
        setNewSecretProvider("custom");
    };

    const safeSecrets = Array.isArray(secrets) ? secrets : [];
    const filteredSecrets = safeSecrets.filter(s =>
        s.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        s.key.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const handleSelectAll = (checked: boolean) => {
        if (checked) {
            setSelectedSecrets(new Set(filteredSecrets.map(s => s.id)));
        } else {
            setSelectedSecrets(new Set());
        }
    };

    const handleSelectOne = (id: string, checked: boolean) => {
        const newSelected = new Set(selectedSecrets);
        if (checked) {
            newSelected.add(id);
        } else {
            newSelected.delete(id);
        }
        setSelectedSecrets(newSelected);
    };

    const isAllSelected = filteredSecrets.length > 0 && selectedSecrets.size === filteredSecrets.length;

    return (
        <div className="space-y-4 h-full flex flex-col">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium">API Keys & Secrets</h3>
                    <p className="text-sm text-muted-foreground">
                        Manage secure credentials for your upstream services.
                    </p>
                </div>
                <div className="flex gap-2">
                    {selectedSecrets.size > 0 && (
                        <div className="flex items-center gap-2 animate-in fade-in slide-in-from-right-4 duration-300">
                             <Button size="sm" variant="destructive" onClick={() => setIsBulkDeleteDialogOpen(true)}>
                                <Trash2 className="mr-2 h-4 w-4" /> Delete ({selectedSecrets.size})
                             </Button>
                        </div>
                    )}
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
            </div>

            <Card className="flex-1 flex flex-col overflow-hidden bg-background/50 backdrop-blur-sm border-muted/50">
                <CardHeader className="p-4 border-b bg-muted/20">
                     <div className="relative">
                        <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            placeholder="Search secrets..."
                            className="pl-8 bg-background max-w-sm"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>
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
                                     <TableRow>
                                         <TableHead className="w-[50px]">
                                             <Checkbox
                                                 checked={isAllSelected}
                                                 onCheckedChange={(checked) => handleSelectAll(!!checked)}
                                                 aria-label="Select all"
                                             />
                                         </TableHead>
                                         <TableHead>Name</TableHead>
                                         <TableHead>Key</TableHead>
                                         <TableHead>Value</TableHead>
                                         <TableHead className="text-right">Actions</TableHead>
                                     </TableRow>
                                 </TableHeader>
                                 <TableBody>
                                     {filteredSecrets.map((secret) => (
                                         <SecretRow
                                             key={secret.id}
                                             secret={secret}
                                             isSelected={selectedSecrets.has(secret.id)}
                                             onSelect={handleSelectOne}
                                             onDelete={handleDeleteSecret}
                                         />
                                     ))}
                                 </TableBody>
                             </Table>
                         )}
                    </ScrollArea>
                </CardContent>
            </Card>

            <AlertDialog open={isBulkDeleteDialogOpen} onOpenChange={setIsBulkDeleteDialogOpen}>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This action cannot be undone. This will permanently delete {selectedSecrets.size} secrets.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction onClick={handleBulkDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">Delete</AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}

/**
 * SecretRow component.
 * @param props - The component props.
 * @param props.secret - The secret property.
 * @param props.isSelected - The isSelected property.
 * @param props.onSelect - The onSelect property.
 * @param props.onDelete - The onDelete property.
 * @returns The rendered component.
 */
function SecretRow({ secret, isSelected, onSelect, onDelete }: {
    secret: SecretDefinition;
    isSelected: boolean;
    onSelect: (id: string, checked: boolean) => void;
    onDelete: (id: string) => void
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
        <TableRow className="group">
             <TableCell>
                 <Checkbox
                     checked={isSelected}
                     onCheckedChange={(checked) => onSelect(secret.id, !!checked)}
                     aria-label={`Select ${secret.name}`}
                 />
             </TableCell>
             <TableCell>
                 <div className="flex items-center gap-2">
                     <div className="bg-primary/10 p-1.5 rounded-full text-primary">
                         <Key className="h-3 w-3" />
                     </div>
                     <div className="flex flex-col">
                         <span className="font-medium text-sm">{secret.name}</span>
                         <Badge variant="outline" className="text-[10px] w-fit font-mono">
                             {secret.provider}
                         </Badge>
                     </div>
                 </div>
             </TableCell>
             <TableCell className="font-mono text-xs text-muted-foreground">
                 {secret.key}
             </TableCell>
             <TableCell>
                 <div className="flex items-center gap-2 bg-muted/50 rounded-md px-2 py-1 border font-mono text-xs w-[200px] justify-between">
                    <span className="truncate">
                        {loading ? <RefreshCw className="h-3 w-3 animate-spin" /> :
                            revealedValue ? revealedValue : "•".repeat(24)}
                    </span>
                     <Button
                         variant="ghost"
                         size="icon"
                         className="h-4 w-4 hover:bg-transparent"
                         onClick={handleReveal}
                         disabled={loading}
                         aria-label={revealedValue ? "Hide secret" : "Show secret"}
                     >
                         {revealedValue ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
                     </Button>
                 </div>
             </TableCell>
             <TableCell className="text-right">
                <div className="flex justify-end gap-1">
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
