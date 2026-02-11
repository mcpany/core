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
import { Trash2, Plus, Info } from "lucide-react";
import { ExportPolicy, ExportPolicy_Action, ExportRule } from "@proto/config/v1/upstream_service";
import { Checkbox } from "@/components/ui/checkbox";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";

interface PolicyEditorProps {
    title: string;
    description: string;
    policy: ExportPolicy | undefined;
    knownItems?: string[];
    onChange: (policy: ExportPolicy) => void;
}

/**
 * A component for editing export policies (allow/deny rules).
 * Allows setting a default action and adding specific exception rules based on regex.
 *
 * @param props - The component props.
 * @param props.title - The title of the policy editor section.
 * @param props.description - A description of what this policy controls.
 * @param props.policy - The current policy object.
 * @param props.knownItems - Optional list of known items (tools, prompts, etc.) to show in a quick select list.
 * @param props.onChange - Callback invoked when the policy is modified.
 * @returns The rendered policy editor component.
 */
export function PolicyEditor({ title, description, policy, knownItems, onChange }: PolicyEditorProps) {
    // Default to empty policy if undefined
    const currentPolicy: ExportPolicy = policy || {
        defaultAction: ExportPolicy_Action.EXPORT, // Default to Allow All usually?
        rules: []
    };

    const handleDefaultActionChange = (val: string) => {
        onChange({
            ...currentPolicy,
            defaultAction: parseInt(val) as ExportPolicy_Action
        });
    };

    const addRule = () => {
        const newRules = [
            ...(currentPolicy.rules || []),
            { nameRegex: "", action: ExportPolicy_Action.UNEXPORT } // Default to "Block specific" if allow all, or "Allow specific" if deny all?
        ];
        onChange({ ...currentPolicy, rules: newRules });
    };

    const updateRule = (index: number, updates: Partial<ExportRule>) => {
        const newRules = [...(currentPolicy.rules || [])];
        newRules[index] = { ...newRules[index], ...updates };
        onChange({ ...currentPolicy, rules: newRules });
    };

    const removeRule = (index: number) => {
        const newRules = [...(currentPolicy.rules || [])];
        newRules.splice(index, 1);
        onChange({ ...currentPolicy, rules: newRules });
    };

    // Helper to determine if an item is effectively exposed
    const isExposed = (name: string) => {
        // Iterate rules in order (First match wins)
        if (currentPolicy.rules) {
            for (const rule of currentPolicy.rules) {
                try {
                    const regex = new RegExp(rule.nameRegex);
                    if (regex.test(name)) {
                        return rule.action === ExportPolicy_Action.EXPORT;
                    }
                } catch (e) {
                    // Ignore invalid regex
                }
            }
        }
        // Fallback
        return currentPolicy.defaultAction === ExportPolicy_Action.EXPORT;
    };

    // Helper to toggle exposure
    const toggleExposed = (name: string, checked: boolean) => {
        const escapedName = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        const regex = `^${escapedName}$`;
        const newRules = [...(currentPolicy.rules || [])];
        const ruleIndex = newRules.findIndex(r => r.nameRegex === regex);

        if (currentPolicy.defaultAction === ExportPolicy_Action.EXPORT) {
            // Default: ALLOW ALL
            if (checked) {
                // User wants to SHOW. Remove any blocking rule.
                if (ruleIndex !== -1) {
                    newRules.splice(ruleIndex, 1);
                }
            } else {
                // User wants to HIDE. Add blocking rule.
                if (ruleIndex === -1) {
                    newRules.unshift({ nameRegex: regex, action: ExportPolicy_Action.UNEXPORT });
                } else {
                    newRules[ruleIndex].action = ExportPolicy_Action.UNEXPORT;
                    // Move to top to ensure priority
                    const rule = newRules.splice(ruleIndex, 1)[0];
                    newRules.unshift(rule);
                }
            }
        } else {
            // Default: DENY ALL
            if (checked) {
                // User wants to SHOW. Add allowing rule.
                if (ruleIndex === -1) {
                    newRules.unshift({ nameRegex: regex, action: ExportPolicy_Action.EXPORT });
                } else {
                    newRules[ruleIndex].action = ExportPolicy_Action.EXPORT;
                    // Move to top to ensure priority
                    const rule = newRules.splice(ruleIndex, 1)[0];
                    newRules.unshift(rule);
                }
            } else {
                // User wants to HIDE. Remove any allowing rule.
                if (ruleIndex !== -1) {
                    newRules.splice(ruleIndex, 1);
                }
            }
        }
        onChange({ ...currentPolicy, rules: newRules });
    };

    return (
        <Card>
            <CardHeader>
                <div className="flex items-center justify-between">
                    <div>
                        <CardTitle className="text-base">{title}</CardTitle>
                        <CardDescription>{description}</CardDescription>
                    </div>
                     <Badge variant="outline" className="font-mono text-xs">
                        {currentPolicy.rules?.length || 0} Custom Rules
                    </Badge>
                </div>
            </CardHeader>
            <CardContent className="space-y-6">
                <div className="flex items-center justify-between bg-muted/20 p-3 rounded-lg border">
                    <Label className="text-sm font-medium">Default Policy</Label>
                    <Select
                        value={currentPolicy.defaultAction.toString()}
                        onValueChange={handleDefaultActionChange}
                    >
                        <SelectTrigger className="w-[200px] bg-background">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value={ExportPolicy_Action.EXPORT.toString()}>Allow All (Whitelist Mode)</SelectItem>
                            <SelectItem value={ExportPolicy_Action.UNEXPORT.toString()}>Deny All (Blacklist Mode)</SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                {knownItems && knownItems.length > 0 && (
                     <div className="space-y-3">
                        <div className="flex items-center gap-2">
                            <Label className="text-sm font-medium">Quick Visibility Control</Label>
                             <div className="group relative">
                                <Info className="h-3 w-3 text-muted-foreground cursor-help" />
                                <div className="absolute left-1/2 -translate-x-1/2 bottom-full mb-2 hidden group-hover:block w-64 bg-popover text-popover-foreground text-xs p-2 rounded border shadow-lg z-50">
                                    Toggle items to override the default policy. Changes here automatically create specific regex rules.
                                </div>
                            </div>
                        </div>
                        <ScrollArea className="h-[200px] rounded-md border p-4 bg-muted/10">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                                {knownItems.map((item) => {
                                    const isVisible = isExposed(item);
                                    return (
                                        <div key={item} className="flex items-center space-x-2 p-1 hover:bg-muted/50 rounded transition-colors">
                                            <Checkbox
                                                id={`item-${item}`}
                                                checked={isVisible}
                                                onCheckedChange={(checked) => toggleExposed(item, checked === true)}
                                            />
                                            <Label
                                                htmlFor={`item-${item}`}
                                                className={`text-sm cursor-pointer flex-1 truncate ${!isVisible && "text-muted-foreground line-through opacity-70"}`}
                                                title={item}
                                            >
                                                {item}
                                            </Label>
                                        </div>
                                    );
                                })}
                            </div>
                        </ScrollArea>
                    </div>
                )}

                <div className="space-y-2 pt-2 border-t">
                    <div className="flex items-center justify-between">
                        <Label className="text-sm font-medium">Advanced Regex Rules</Label>
                        <Button variant="outline" size="sm" onClick={addRule}>
                            <Plus className="h-3 w-3 mr-1" /> Add Rule
                        </Button>
                    </div>

                    {(!currentPolicy.rules || currentPolicy.rules.length === 0) && (
                        <div className="text-sm text-muted-foreground italic p-4 border border-dashed rounded-md text-center bg-muted/5">
                            No custom regex rules defined.
                        </div>
                    )}

                    {currentPolicy.rules?.map((rule, index) => (
                        <div key={index} className="flex gap-2 items-start animate-in fade-in slide-in-from-top-1 duration-200">
                            <div className="flex-1">
                                <Input
                                    placeholder="Name Regex (e.g. ^delete_.*)"
                                    value={rule.nameRegex}
                                    onChange={(e) => updateRule(index, { nameRegex: e.target.value })}
                                    className="font-mono text-sm"
                                />
                            </div>
                            <Select
                                value={rule.action.toString()}
                                onValueChange={(val) => updateRule(index, { action: parseInt(val) as ExportPolicy_Action })}
                            >
                                <SelectTrigger className="w-[120px]">
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={ExportPolicy_Action.EXPORT.toString()}>Allow</SelectItem>
                                    <SelectItem value={ExportPolicy_Action.UNEXPORT.toString()}>Deny</SelectItem>
                                </SelectContent>
                            </Select>
                            <Button variant="ghost" size="icon" onClick={() => removeRule(index)} className="text-destructive hover:text-destructive hover:bg-destructive/10">
                                <Trash2 className="h-4 w-4" />
                            </Button>
                        </div>
                    ))}
                </div>
            </CardContent>
        </Card>
    );
}
