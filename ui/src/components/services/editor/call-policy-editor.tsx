/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Trash2, Plus, ShieldAlert, ShieldCheck } from "lucide-react";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";

// Re-define types locally to avoid complex import issues if protos aren't perfectly aligned
// This matches `proto/config/v1/upstream_service.proto`
export enum CallPolicyAction {
    ALLOW = 0,
    DENY = 1,
    SAVE_CACHE = 2,
    DELETE_CACHE = 3
}

export interface CallPolicyRule {
    action: CallPolicyAction;
    nameRegex?: string;
    argumentRegex?: string;
    urlRegex?: string;
    callIdRegex?: string;
}

export interface CallPolicy {
    defaultAction: CallPolicyAction;
    rules: CallPolicyRule[];
}

interface CallPolicyEditorProps {
    policies: CallPolicy[] | undefined;
    onChange: (policies: CallPolicy[]) => void;
}

const ACTION_LABELS: Record<number, string> = {
    [CallPolicyAction.ALLOW]: "Allow",
    [CallPolicyAction.DENY]: "Deny",
    [CallPolicyAction.SAVE_CACHE]: "Save Cache",
    [CallPolicyAction.DELETE_CACHE]: "Delete Cache",
};

/**
 * A component for editing call policies (security rules).
 * Allows managing a list of policies, each with a default action and specific regex-based rules.
 */
export function CallPolicyEditor({ policies, onChange }: CallPolicyEditorProps) {
    const currentPolicies = policies || [];

    const addPolicy = () => {
        onChange([
            ...currentPolicies,
            { defaultAction: CallPolicyAction.DENY, rules: [] }
        ]);
    };

    const updatePolicy = (index: number, updates: Partial<CallPolicy>) => {
        const newPolicies = [...currentPolicies];
        newPolicies[index] = { ...newPolicies[index], ...updates };
        onChange(newPolicies);
    };

    const removePolicy = (index: number) => {
        const newPolicies = [...currentPolicies];
        newPolicies.splice(index, 1);
        onChange(newPolicies);
    };

    const addRule = (policyIndex: number) => {
        const policy = currentPolicies[policyIndex];
        const newRules = [
            ...(policy.rules || []),
            { action: CallPolicyAction.ALLOW, nameRegex: "" }
        ];
        updatePolicy(policyIndex, { rules: newRules });
    };

    const updateRule = (policyIndex: number, ruleIndex: number, updates: Partial<CallPolicyRule>) => {
        const policy = currentPolicies[policyIndex];
        const newRules = [...(policy.rules || [])];
        newRules[ruleIndex] = { ...newRules[ruleIndex], ...updates };
        updatePolicy(policyIndex, { rules: newRules });
    };

    const removeRule = (policyIndex: number, ruleIndex: number) => {
        const policy = currentPolicies[policyIndex];
        const newRules = [...(policy.rules || [])];
        newRules.splice(ruleIndex, 1);
        updatePolicy(policyIndex, { rules: newRules });
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium flex items-center gap-2">
                        <ShieldAlert className="h-5 w-5 text-primary" />
                        Security & Access Control
                    </h3>
                    <p className="text-sm text-muted-foreground">
                        Define policies to allow or deny specific tool calls.
                    </p>
                </div>
                <Button onClick={addPolicy} variant="outline" size="sm">
                    <Plus className="mr-2 h-4 w-4" /> Add Policy
                </Button>
            </div>

            {currentPolicies.length === 0 && (
                <Card className="border-dashed">
                    <CardContent className="flex flex-col items-center justify-center p-6 text-muted-foreground italic">
                        No policies configured. All calls are allowed by default unless restricted.
                        <Button onClick={addPolicy} variant="link" className="mt-2">
                            Create a Policy
                        </Button>
                    </CardContent>
                </Card>
            )}

            <Accordion type="multiple" className="w-full space-y-4">
                {currentPolicies.map((policy, policyIndex) => (
                    <AccordionItem key={policyIndex} value={`policy-${policyIndex}`} className="border rounded-lg px-4">
                        <AccordionTrigger className="hover:no-underline">
                            <div className="flex items-center gap-4 w-full">
                                <span className="font-semibold text-sm">Policy #{policyIndex + 1}</span>
                                <Badge variant={policy.defaultAction === CallPolicyAction.DENY ? "destructive" : "outline"}>
                                    Default: {ACTION_LABELS[policy.defaultAction]}
                                </Badge>
                                <span className="text-xs text-muted-foreground ml-auto mr-4">
                                    {policy.rules?.length || 0} Rules
                                </span>
                            </div>
                        </AccordionTrigger>
                        <AccordionContent className="pt-4 pb-4 space-y-4 border-t mt-2">
                             <div className="flex items-center justify-between bg-muted/30 p-3 rounded-md">
                                <Label className="text-sm font-medium">Default Action (Fallback)</Label>
                                <Select
                                    value={policy.defaultAction.toString()}
                                    onValueChange={(val) => updatePolicy(policyIndex, { defaultAction: parseInt(val) })}
                                >
                                    <SelectTrigger className="w-[200px]">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value={CallPolicyAction.ALLOW.toString()}>Allow All</SelectItem>
                                        <SelectItem value={CallPolicyAction.DENY.toString()}>Deny All</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                             <div className="space-y-2">
                                <div className="flex items-center justify-between">
                                    <Label className="text-sm font-medium">Rules (First Match Wins)</Label>
                                    <Button variant="ghost" size="sm" onClick={() => addRule(policyIndex)}>
                                        <Plus className="h-3 w-3 mr-1" /> Add Rule
                                    </Button>
                                </div>

                                <div className="space-y-3">
                                    {(!policy.rules || policy.rules.length === 0) && (
                                        <div className="text-sm text-muted-foreground italic p-2 border border-dashed rounded-md text-center">
                                            No exception rules defined.
                                        </div>
                                    )}

                                    {policy.rules?.map((rule, ruleIndex) => (
                                        <Card key={ruleIndex} className="bg-muted/10 border-l-4 border-l-primary/50">
                                            <CardContent className="p-3 grid gap-3">
                                                <div className="flex justify-between items-start gap-2">
                                                    <div className="flex-1 grid grid-cols-2 gap-2">
                                                         <div className="col-span-2 sm:col-span-1">
                                                            <Label className="text-[10px] text-muted-foreground">Action</Label>
                                                            <Select
                                                                value={rule.action.toString()}
                                                                onValueChange={(val) => updateRule(policyIndex, ruleIndex, { action: parseInt(val) })}
                                                            >
                                                                <SelectTrigger className="h-8 text-xs">
                                                                    <SelectValue />
                                                                </SelectTrigger>
                                                                <SelectContent>
                                                                    <SelectItem value={CallPolicyAction.ALLOW.toString()}>Allow</SelectItem>
                                                                    <SelectItem value={CallPolicyAction.DENY.toString()}>Deny</SelectItem>
                                                                    <SelectItem value={CallPolicyAction.SAVE_CACHE.toString()}>Save Cache</SelectItem>
                                                                    <SelectItem value={CallPolicyAction.DELETE_CACHE.toString()}>Delete Cache</SelectItem>
                                                                </SelectContent>
                                                            </Select>
                                                        </div>
                                                        <div className="col-span-2 sm:col-span-1">
                                                            <Label className="text-[10px] text-muted-foreground">Call ID Regex</Label>
                                                            <Input
                                                                className="h-8 text-xs font-mono"
                                                                placeholder=".*"
                                                                value={rule.callIdRegex || ""}
                                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { callIdRegex: e.target.value })}
                                                            />
                                                        </div>
                                                        <div className="col-span-2">
                                                            <Label className="text-[10px] text-muted-foreground">Name Regex</Label>
                                                            <Input
                                                                className="h-8 text-xs font-mono"
                                                                placeholder="e.g. ^git.*"
                                                                value={rule.nameRegex || ""}
                                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { nameRegex: e.target.value })}
                                                            />
                                                        </div>
                                                        <div className="col-span-2">
                                                             <Label className="text-[10px] text-muted-foreground">Argument Regex (JSON)</Label>
                                                            <Input
                                                                className="h-8 text-xs font-mono"
                                                                placeholder='e.g. "path": "/etc/.*"'
                                                                value={rule.argumentRegex || ""}
                                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { argumentRegex: e.target.value })}
                                                            />
                                                        </div>
                                                    </div>
                                                    <Button
                                                        variant="ghost"
                                                        size="icon"
                                                        onClick={() => removeRule(policyIndex, ruleIndex)}
                                                        className="h-6 w-6 text-muted-foreground hover:text-destructive shrink-0 mt-4"
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            </CardContent>
                                        </Card>
                                    ))}
                                </div>
                             </div>

                            <div className="flex justify-end pt-2">
                                <Button
                                    variant="destructive"
                                    size="sm"
                                    onClick={() => removePolicy(policyIndex)}
                                    className="opacity-80 hover:opacity-100"
                                >
                                    <Trash2 className="mr-2 h-4 w-4" /> Delete Policy
                                </Button>
                            </div>
                        </AccordionContent>
                    </AccordionItem>
                ))}
            </Accordion>
        </div>
    );
}
