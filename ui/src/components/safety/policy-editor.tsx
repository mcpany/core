/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { CallPolicy, CallPolicyRule, CallPolicy_Action } from "@proto/config/v1/upstream_service";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Shield, ShieldAlert, ShieldCheck, Plus, Trash2, Edit2, Info } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/hooks/use-toast";

interface PolicyEditorProps {
    policies: CallPolicy[];
    onUpdate: (policies: CallPolicy[]) => void;
}

const ACTION_LABELS: Record<number, string> = {
    [CallPolicy_Action.ALLOW]: "Allow",
    [CallPolicy_Action.DENY]: "Deny",
    [CallPolicy_Action.SAVE_CACHE]: "Save Cache",
    [CallPolicy_Action.DELETE_CACHE]: "Delete Cache",
};

const ACTION_COLORS: Record<number, "default" | "destructive" | "secondary" | "outline"> = {
    [CallPolicy_Action.ALLOW]: "default", // Green-ish usually, but default works
    [CallPolicy_Action.DENY]: "destructive",
    [CallPolicy_Action.SAVE_CACHE]: "secondary",
    [CallPolicy_Action.DELETE_CACHE]: "secondary",
};

/**
 * PolicyEditor allows viewing and editing of advanced call policies.
 * It provides a UI to manage default actions and specific rules for call interception and handling.
 *
 * @param props - The component props.
 * @param props.policies - The list of current policies.
 * @param props.onUpdate - Callback when policies are updated.
 * @returns The rendered PolicyEditor component.
 */
export function PolicyEditor({ policies = [], onUpdate }: PolicyEditorProps) {
    const { toast } = useToast();
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [currentPolicy, setCurrentPolicy] = useState<CallPolicy>({
        defaultAction: CallPolicy_Action.DENY,
        rules: []
    });

    const handleOpenCreate = () => {
        setEditingIndex(null);
        setCurrentPolicy({
            defaultAction: CallPolicy_Action.DENY,
            rules: []
        });
        setIsDialogOpen(true);
    };

    const handleEdit = (index: number) => {
        setEditingIndex(index);
        // Deep copy to avoid mutating props directly
        setCurrentPolicy(JSON.parse(JSON.stringify(policies[index])));
        setIsDialogOpen(true);
    };

    const handleDelete = (index: number) => {
        const newPolicies = [...policies];
        newPolicies.splice(index, 1);
        onUpdate(newPolicies);
        toast({ title: "Policy Deleted", description: "The policy has been removed." });
    };

    const handleSave = () => {
        const newPolicies = [...policies];
        if (editingIndex !== null) {
            newPolicies[editingIndex] = currentPolicy;
        } else {
            newPolicies.push(currentPolicy);
        }
        onUpdate(newPolicies);
        setIsDialogOpen(false);
        toast({ title: "Policy Saved", description: "The policy has been updated." });
    };

    const addRule = () => {
        setCurrentPolicy({
            ...currentPolicy,
            rules: [
                ...currentPolicy.rules,
                {
                    action: CallPolicy_Action.ALLOW,
                    nameRegex: "",
                    argumentRegex: "",
                    urlRegex: "",
                    callIdRegex: ""
                }
            ]
        });
    };

    const updateRule = (index: number, field: keyof CallPolicyRule, value: any) => {
        const newRules = [...currentPolicy.rules];
        newRules[index] = { ...newRules[index], [field]: value };
        setCurrentPolicy({ ...currentPolicy, rules: newRules });
    };

    const deleteRule = (index: number) => {
        const newRules = [...currentPolicy.rules];
        newRules.splice(index, 1);
        setCurrentPolicy({ ...currentPolicy, rules: newRules });
    };

    return (
        <Card className="border-l-4 border-l-primary/20">
            <CardHeader className="flex flex-row items-center justify-between">
                <div>
                    <CardTitle className="flex items-center gap-2"><Shield className="h-5 w-5" /> Advanced Call Policies</CardTitle>
                    <CardDescription>
                        Define granular access control rules based on tool names, arguments, and more.
                        Policies are evaluated in order.
                    </CardDescription>
                </div>
                <Button onClick={handleOpenCreate} size="sm">
                    <Plus className="mr-2 h-4 w-4" /> Add Policy
                </Button>
            </CardHeader>
            <CardContent>
                {policies.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground border-2 border-dashed rounded-lg">
                        No policies configured. All calls are governed by global defaults (usually allowed).
                    </div>
                ) : (
                    <div className="space-y-4">
                        {policies.map((policy, idx) => (
                            <Card key={idx} className="bg-muted/10">
                                <CardContent className="p-4 flex items-center justify-between">
                                    <div className="flex items-center gap-4">
                                        <div className="flex flex-col items-center min-w-[60px]">
                                            <span className="text-xs font-mono uppercase text-muted-foreground mb-1">Default</span>
                                            {policy.defaultAction === CallPolicy_Action.ALLOW ? (
                                                <ShieldCheck className="h-6 w-6 text-green-500" />
                                            ) : (
                                                <ShieldAlert className="h-6 w-6 text-red-500" />
                                            )}
                                            <span className="text-xs font-bold mt-1">{ACTION_LABELS[policy.defaultAction]}</span>
                                        </div>
                                        <Separator orientation="vertical" className="h-10" />
                                        <div>
                                            <div className="font-medium">{policy.rules.length} Rule{policy.rules.length !== 1 ? 's' : ''}</div>
                                            <div className="text-sm text-muted-foreground">
                                                {policy.rules.slice(0, 2).map((r, i) => (
                                                    <div key={i} className="flex items-center gap-1 mt-1">
                                                        <Badge variant="outline" className="text-[10px] py-0 h-4">{ACTION_LABELS[r.action]}</Badge>
                                                        <span className="font-mono text-xs truncate max-w-[200px]">
                                                            {r.nameRegex ? `Name: /${r.nameRegex}/` : r.argumentRegex ? `Args: /${r.argumentRegex}/` : "Match All"}
                                                        </span>
                                                    </div>
                                                ))}
                                                {policy.rules.length > 2 && <div className="text-xs mt-1">+{policy.rules.length - 2} more...</div>}
                                            </div>
                                        </div>
                                    </div>
                                    <div className="flex gap-2">
                                        <Button variant="ghost" size="icon" onClick={() => handleEdit(idx)}>
                                            <Edit2 className="h-4 w-4" />
                                        </Button>
                                        <Button variant="ghost" size="icon" className="text-destructive hover:text-destructive" onClick={() => handleDelete(idx)}>
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                )}
            </CardContent>

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="max-w-3xl max-h-[80vh] flex flex-col">
                    <DialogHeader>
                        <DialogTitle>{editingIndex !== null ? "Edit Policy" : "Create Policy"}</DialogTitle>
                        <DialogDescription>
                            Configure the default action and specific matching rules.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="flex-1 overflow-y-auto py-4 space-y-6">
                        <div className="grid gap-2">
                            <Label>Default Action</Label>
                            <Select
                                value={currentPolicy.defaultAction.toString()}
                                onValueChange={(v) => setCurrentPolicy({ ...currentPolicy, defaultAction: parseInt(v) })}
                            >
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={CallPolicy_Action.ALLOW.toString()}>Allow</SelectItem>
                                    <SelectItem value={CallPolicy_Action.DENY.toString()}>Deny</SelectItem>
                                </SelectContent>
                            </Select>
                            <p className="text-xs text-muted-foreground">
                                Applied when no rules match the request.
                            </p>
                        </div>

                        <Separator />

                        <div className="space-y-4">
                            <div className="flex items-center justify-between">
                                <Label>Rules (Evaluated Top-Down)</Label>
                                <Button variant="outline" size="sm" onClick={addRule}>
                                    <Plus className="mr-2 h-3 w-3" /> Add Rule
                                </Button>
                            </div>

                            {currentPolicy.rules.length === 0 && (
                                <div className="text-center py-4 text-sm text-muted-foreground border border-dashed rounded">
                                    No rules defined.
                                </div>
                            )}

                            {currentPolicy.rules.map((rule, idx) => (
                                <div key={idx} className="border rounded-md p-4 space-y-4 relative group bg-muted/5">
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity text-destructive hover:text-destructive hover:bg-destructive/10 h-6 w-6"
                                        onClick={() => deleteRule(idx)}
                                    >
                                        <Trash2 className="h-3 w-3" />
                                    </Button>

                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="space-y-2">
                                            <Label className="text-xs">Action</Label>
                                            <Select
                                                value={rule.action.toString()}
                                                onValueChange={(v) => updateRule(idx, "action", parseInt(v))}
                                            >
                                                <SelectTrigger className="h-8">
                                                    <SelectValue />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value={CallPolicy_Action.ALLOW.toString()}>Allow</SelectItem>
                                                    <SelectItem value={CallPolicy_Action.DENY.toString()}>Deny</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>
                                        <div className="space-y-2">
                                            <Label className="text-xs">Tool Name Regex</Label>
                                            <Input
                                                className="h-8 font-mono text-xs"
                                                placeholder="e.g. ^git.*"
                                                value={rule.nameRegex}
                                                onChange={(e) => updateRule(idx, "nameRegex", e.target.value)}
                                            />
                                        </div>
                                    </div>
                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="space-y-2">
                                            <Label className="text-xs">Argument Regex (JSON)</Label>
                                            <Input
                                                className="h-8 font-mono text-xs"
                                                placeholder='e.g. "force": true'
                                                value={rule.argumentRegex}
                                                onChange={(e) => updateRule(idx, "argumentRegex", e.target.value)}
                                            />
                                        </div>
                                        <div className="space-y-2">
                                            <Label className="text-xs">Call ID Regex</Label>
                                            <Input
                                                className="h-8 font-mono text-xs"
                                                placeholder="e.g. .*"
                                                value={rule.callIdRegex}
                                                onChange={(e) => updateRule(idx, "callIdRegex", e.target.value)}
                                            />
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    <DialogFooter className="mt-4">
                        <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                        <Button onClick={handleSave}>Save Policy</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </Card>
    );
}
