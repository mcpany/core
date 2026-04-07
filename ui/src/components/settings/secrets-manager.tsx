/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



import { useState, useEffect, useMemo, useCallback } from "react";
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
import { Checkbox } from "@/components/ui/checkbox";
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
import { cn } from "@/lib/utils";

/**
 * Intent: Document SecretsManager
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * SecretsManager component.
 * @returns The rendered component.
 */
export function SecretsManager() {
    const [secrets, setSecrets] = useState<SecretDefinition[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState("");
    const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
    const [selected, setSelected] = useState<Set<string>>(new Set());
    const { toast } = useToast();

    // Form state
    const [newSecretName, setNewSecretName] = useState("");
    const [newSecretKey, setNewSecretKey] = useState("");
    const [newSecretValue, setNewSecretValue] = useState("");
    const [newSecretProvider, setNewSecretProvider] = useState<string>("custom");

    useEffect(() => {
        loadSecrets();
    }, []);

    // Reset selection when secrets list changes
    useEffect(() => {
        setSelected(new Set());
    }, [secrets]);

    const handleSelectAll = useCallback((checked: boolean, filteredSecretsList: SecretDefinition[]) => {
        if (checked) {
            setSelected(new Set(filteredSecretsList.map(s => s.id)));
        } else {
            setSelected(new Set());
        }
    }, []);

    const handleSelectOne = useCallback((id: string, checked: boolean) => {
        setSelected(prev => {
            const newSelected = new Set(prev);
            if (checked) {
                newSelected.add(id);
            } else {
                newSelected.delete(id);
            }
            return newSelected;
        });
    }, []);

    const handleBulkDelete = async () => {
        const selectedIds = Array.from(selected);
        if (!confirm(`Are you sure you want to delete ${selectedIds.length} secrets?`)) return;

        try {
            await Promise.all(selectedIds.map(id => apiClient.deleteSecret(id)));
            toast({
                title: "Secrets Deleted",
                description: `${selectedIds.length} secrets have been removed.`
            });
            setSelected(new Set());
            loadSecrets();
        } catch (e) {
            console.error("Failed to delete secrets", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: "Failed to delete some secrets."
            });
            loadSecrets();
        }
    };

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
            loadSecrets();
        } catch (_error) {
            toast({
                title: "Error",
                description: "Failed to delete secret.",
                variant: "destructive",
            });
        }
    };

    const resetForm = () => {
        setNewSecretName("");
        setNewSecretKey("");
        setNewSecretValue("");
        setNewSecretProvider("custom");
    };

    const safeSecrets = Array.isArray(secrets) ? secrets : [];

    // ⚡ BOLT: Memoize filtered secrets and avoid calling toLowerCase() repeatedly inside the filter loop.
    // Randomized Selection from Top 5 High-Impact Targets
    const filteredSecrets = useMemo(() => {
        const query = searchQuery.toLowerCase();
        if (!query) return safeSecrets;

        return safeSecrets.filter(s =>
            s.name.toLowerCase().includes(query) ||
            s.key.toLowerCase().includes(query)
        );
    }, [safeSecrets, searchQuery]);

    const isAllSelected = filteredSecrets.length > 0 && selected.size === filteredSecrets.length;

    return (
        <div className="space-y-4 h-full flex flex-col">
            {selected.size > 0 && (
                <div className="flex items-center gap-2 p-2 bg-muted/40 rounded-md animate-in fade-in slide-in-from-top-1 duration-200 sticky top-0 z-10 backdrop-blur-md border">
                    <span className="text-sm text-muted-foreground mr-2 font-medium px-2">{selected.size} selected</span>
                    <div className="h-4 w-px bg-border mx-1" />
                    <Button size="sm" variant="ghost" onClick={handleBulkDelete} className="h-8 text-red-600 hover:text-red-700 hover:bg-red-100 dark:hover:bg-red-900/20">
                        <Trash2 className="mr-2 h-4 w-4" /> Delete Selected
                    </Button>
                </div>
            )}

            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium">API Keys & Secrets</h3>
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
                <CardHeader className="p-4 border-b bg-muted/20 flex flex-row items-center justify-between">
                    <div className="flex items-center gap-4 w-full">
                        <div className="flex items-center space-x-2">
                            <Checkbox
                                id="select-all"
                                checked={isAllSelected}
                                onCheckedChange={(checked) => handleSelectAll(!!checked, filteredSecrets)}
                                aria-label="Select all"
                            />
                            <Label htmlFor="select-all" className="text-sm font-medium cursor-pointer">
                                Select All
                            </Label>
                        </div>
                        <div className="relative flex-1 max-w-sm">
                            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search secrets..."
                                className="pl-8 bg-background"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                            />
                        </div>
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
                            <div className="divide-y">
                                {filteredSecrets.map((secret) => (
                                    <SecretItem
                                        key={secret.id}
                                        secret={secret}
                                        onDelete={handleDeleteSecret}
                                        isSelected={selected.has(secret.id)}
                                        onSelect={(checked) => handleSelectOne(secret.id, checked)}
                                    />
                                ))}
                            </div>
                        )}
                    </ScrollArea>
                </CardContent>
            </Card>
        </div>
    );
}

/**
 * SecretItem component.
 * @param props - The component props.
 * @param props.secret - The secret property.
 * @param props.onDelete - The onDelete property.
 * @param props.isSelected - Whether the item is selected.
 * @param props.onSelect - Callback when selection changes.
 * @returns The rendered component.
 */
function SecretItem({ secret, onDelete, isSelected, onSelect }: { secret: SecretDefinition; onDelete: (id: string) => void; isSelected: boolean; onSelect: (checked: boolean) => void }) {
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
        <div className={cn("flex items-center justify-between p-4 hover:bg-muted/30 transition-colors group", isSelected && "bg-muted/50")}>
            <div className="flex items-center gap-4">
                <Checkbox
                    checked={isSelected}
                    onCheckedChange={(checked) => onSelect(!!checked)}
                    aria-label={`Select ${secret.name}`}
                />
                <div className="bg-primary/10 p-2 rounded-full text-primary">
                    <Key className="h-4 w-4" />
                </div>
                <div>
                    <div className="flex items-center gap-2">
                        <h4 className="font-medium text-sm">{secret.name}</h4>
                        <Badge variant="outline" className="text-[10px] h-5 font-mono">
                            {secret.provider}
                        </Badge>
                    </div>
                    <div className="text-xs text-muted-foreground font-mono mt-1">
                        {secret.key}
                    </div>
                </div>
            </div>

            <div className="flex items-center gap-2">
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
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleCopy} disabled={loading} aria-label="Copy secret">
                    <Copy className="h-4 w-4 text-muted-foreground" />
                </Button>
                <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive/70 hover:text-destructive hover:bg-destructive/10" onClick={() => onDelete(secret.id)} aria-label="Delete secret">
                    <Trash2 className="h-4 w-4" />
                </Button>
            </div>
        </div>
    );
}
