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
import { Trash2, Plus, ShieldAlert, ShieldCheck, Database, Eraser } from "lucide-react";
import { CallPolicy, CallPolicy_Action, CallPolicyRule } from "@proto/config/v1/upstream_service";
import { Badge } from "@/components/ui/badge";

interface CallPolicyEditorProps {
    policies: CallPolicy[] | undefined;
    onChange: (policies: CallPolicy[]) => void;
}

const getActionIcon = (action: CallPolicy_Action) => {
    switch (action) {
        case CallPolicy_Action.ALLOW: return <ShieldCheck className="h-4 w-4 text-green-500" />;
        case CallPolicy_Action.DENY: return <ShieldAlert className="h-4 w-4 text-red-500" />;
        case CallPolicy_Action.SAVE_CACHE: return <Database className="h-4 w-4 text-blue-500" />;
        case CallPolicy_Action.DELETE_CACHE: return <Eraser className="h-4 w-4 text-amber-500" />;
        default: return <ShieldCheck className="h-4 w-4" />;
    }
};

const getActionLabel = (action: CallPolicy_Action) => {
    switch (action) {
        case CallPolicy_Action.ALLOW: return "Allow";
        case CallPolicy_Action.DENY: return "Deny";
        case CallPolicy_Action.SAVE_CACHE: return "Save Cache";
        case CallPolicy_Action.DELETE_CACHE: return "Delete Cache";
        default: return "Unknown";
    }
};

/**
 * A component for editing runtime call policies (firewall rules).
 *
 * @param props - The component props.
 * @param props.policies - The current list of call policies.
 * @param props.onChange - Callback invoked when the policies are modified.
 * @returns The rendered call policy editor component.
 */
export function CallPolicyEditor({ policies, onChange }: CallPolicyEditorProps) {
    const currentPolicies = policies || [];

    const addPolicy = () => {
        const newPolicy: CallPolicy = {
            defaultAction: CallPolicy_Action.ALLOW,
            rules: []
        };
        onChange([...currentPolicies, newPolicy]);
    };

    const removePolicy = (index: number) => {
        const newPolicies = [...currentPolicies];
        newPolicies.splice(index, 1);
        onChange(newPolicies);
    };

    const updatePolicy = (index: number, updates: Partial<CallPolicy>) => {
        const newPolicies = [...currentPolicies];
        newPolicies[index] = { ...newPolicies[index], ...updates };
        onChange(newPolicies);
    };

    const addRule = (policyIndex: number) => {
        const policy = currentPolicies[policyIndex];
        const newRules = [
            ...(policy.rules || []),
            {
                action: CallPolicy_Action.DENY,
                nameRegex: "",
                argumentRegex: "",
                urlRegex: "",
                callIdRegex: ""
            }
        ];
        updatePolicy(policyIndex, { rules: newRules });
    };

    const removeRule = (policyIndex: number, ruleIndex: number) => {
        const policy = currentPolicies[policyIndex];
        const newRules = [...(policy.rules || [])];
        newRules.splice(ruleIndex, 1);
        updatePolicy(policyIndex, { rules: newRules });
    };

    const updateRule = (policyIndex: number, ruleIndex: number, updates: Partial<CallPolicyRule>) => {
        const policy = currentPolicies[policyIndex];
        const newRules = [...(policy.rules || [])];
        newRules[ruleIndex] = { ...newRules[ruleIndex], ...updates };
        updatePolicy(policyIndex, { rules: newRules });
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium">Runtime Security Policies</h3>
                    <p className="text-sm text-muted-foreground">
                        Define firewall rules to Allow or Deny specific tool calls based on name, arguments, or ID.
                        Policies are evaluated in order.
                    </p>
                </div>
                <Button onClick={addPolicy}>
                    <Plus className="mr-2 h-4 w-4" /> Add Policy
                </Button>
            </div>

            {currentPolicies.length === 0 && (
                <div className="text-center p-8 border border-dashed rounded-lg bg-muted/20 text-muted-foreground">
                    No active security policies. All calls are allowed by default unless restricted by global settings.
                </div>
            )}

            {currentPolicies.map((policy, policyIndex) => (
                <Card key={policyIndex} className="relative border-l-4 border-l-primary/50">
                    <Button
                        variant="ghost"
                        size="icon"
                        className="absolute right-2 top-2 text-muted-foreground hover:text-destructive"
                        onClick={() => removePolicy(policyIndex)}
                    >
                        <Trash2 className="h-4 w-4" />
                    </Button>

                    <CardHeader className="pb-2">
                        <CardTitle className="text-base flex items-center gap-2">
                            Policy #{policyIndex + 1}
                            <Badge variant="outline" className="font-normal text-xs ml-2">
                                {(policy.rules?.length || 0)} Rules
                            </Badge>
                        </CardTitle>
                        <CardDescription>
                            Evaluates rules top-to-bottom. If no rule matches, the default action is applied.
                        </CardDescription>
                    </CardHeader>

                    <CardContent className="space-y-4">
                        <div className="flex items-center gap-4 bg-muted/30 p-3 rounded-md border">
                            <Label className="text-sm font-medium whitespace-nowrap">Default Action:</Label>
                            <Select
                                value={policy.defaultAction.toString()}
                                onValueChange={(val) => updatePolicy(policyIndex, { defaultAction: parseInt(val) as CallPolicy_Action })}
                            >
                                <SelectTrigger className="w-[180px] h-8">
                                    <div className="flex items-center gap-2">
                                        {getActionIcon(policy.defaultAction)}
                                        <SelectValue />
                                    </div>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={CallPolicy_Action.ALLOW.toString()}>Allow All</SelectItem>
                                    <SelectItem value={CallPolicy_Action.DENY.toString()}>Deny All</SelectItem>
                                    <SelectItem value={CallPolicy_Action.SAVE_CACHE.toString()}>Save Cache</SelectItem>
                                    <SelectItem value={CallPolicy_Action.DELETE_CACHE.toString()}>Delete Cache</SelectItem>
                                </SelectContent>
                            </Select>
                            <span className="text-xs text-muted-foreground">
                                Applied when no rules match.
                            </span>
                        </div>

                        <div className="space-y-2">
                            <div className="flex items-center justify-between">
                                <Label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Rules</Label>
                                <Button variant="outline" size="sm" onClick={() => addRule(policyIndex)} className="h-7 text-xs">
                                    <Plus className="h-3 w-3 mr-1" /> Add Rule
                                </Button>
                            </div>

                            {(!policy.rules || policy.rules.length === 0) && (
                                <div className="text-sm text-muted-foreground italic p-4 text-center">
                                    No specific rules defined.
                                </div>
                            )}

                            {policy.rules?.map((rule, ruleIndex) => (
                                <div key={ruleIndex} className="grid gap-2 p-3 border rounded-md bg-card/50 hover:bg-card transition-colors relative group">
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        onClick={() => removeRule(policyIndex, ruleIndex)}
                                        className="absolute right-1 top-1 h-6 w-6 text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity"
                                    >
                                        <Trash2 className="h-3 w-3" />
                                    </Button>

                                    <div className="flex items-center gap-2 mb-2">
                                        <span className="text-xs font-mono text-muted-foreground w-6 text-right">{ruleIndex + 1}.</span>
                                        <Select
                                            value={rule.action.toString()}
                                            onValueChange={(val) => updateRule(policyIndex, ruleIndex, { action: parseInt(val) as CallPolicy_Action })}
                                        >
                                            <SelectTrigger className="w-[140px] h-8 bg-background">
                                                <div className="flex items-center gap-2">
                                                    {getActionIcon(rule.action)}
                                                    <span>{getActionLabel(rule.action)}</span>
                                                </div>
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value={CallPolicy_Action.ALLOW.toString()}>Allow</SelectItem>
                                                <SelectItem value={CallPolicy_Action.DENY.toString()}>Deny</SelectItem>
                                                <SelectItem value={CallPolicy_Action.SAVE_CACHE.toString()}>Save Cache</SelectItem>
                                                <SelectItem value={CallPolicy_Action.DELETE_CACHE.toString()}>Delete Cache</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <span className="text-xs text-muted-foreground">if...</span>
                                    </div>

                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-2 pl-8">
                                        <div className="space-y-1">
                                            <Label className="text-[10px] text-muted-foreground">Tool Name Regex</Label>
                                            <Input
                                                placeholder=".*"
                                                value={rule.nameRegex}
                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { nameRegex: e.target.value })}
                                                className="h-8 font-mono text-xs"
                                            />
                                        </div>
                                        <div className="space-y-1">
                                            <Label className="text-[10px] text-muted-foreground">Arguments Regex (JSON)</Label>
                                            <Input
                                                placeholder=".*"
                                                value={rule.argumentRegex}
                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { argumentRegex: e.target.value })}
                                                className="h-8 font-mono text-xs"
                                            />
                                        </div>
                                         <div className="space-y-1">
                                            <Label className="text-[10px] text-muted-foreground">URL/Path Regex</Label>
                                            <Input
                                                placeholder=".*"
                                                value={rule.urlRegex}
                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { urlRegex: e.target.value })}
                                                className="h-8 font-mono text-xs"
                                            />
                                        </div>
                                         <div className="space-y-1">
                                            <Label className="text-[10px] text-muted-foreground">Call ID Regex</Label>
                                            <Input
                                                placeholder=".*"
                                                value={rule.callIdRegex}
                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { callIdRegex: e.target.value })}
                                                className="h-8 font-mono text-xs"
                                            />
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </CardContent>
                </Card>
            ))}
        </div>
    );
}
