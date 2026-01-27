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
import { Trash2, Plus } from "lucide-react";
import { ExportPolicy, ExportPolicy_Action, ExportRule } from "@proto/config/v1/upstream_service";

interface PolicyEditorProps {
    title: string;
    description: string;
    policy: ExportPolicy | undefined;
    onChange: (policy: ExportPolicy) => void;
}

export function PolicyEditor({ title, description, policy, onChange }: PolicyEditorProps) {
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

    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-base">{title}</CardTitle>
                <CardDescription>{description}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="flex items-center justify-between">
                    <Label className="text-sm font-medium">Default Action</Label>
                    <Select
                        value={currentPolicy.defaultAction.toString()}
                        onValueChange={handleDefaultActionChange}
                    >
                        <SelectTrigger className="w-[200px]">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value={ExportPolicy_Action.EXPORT.toString()}>Allow All (Export)</SelectItem>
                            <SelectItem value={ExportPolicy_Action.UNEXPORT.toString()}>Deny All (Unexport)</SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                <div className="space-y-2">
                    <div className="flex items-center justify-between">
                        <Label className="text-sm font-medium">Exceptions (Rules)</Label>
                        <Button variant="outline" size="sm" onClick={addRule}>
                            <Plus className="h-3 w-3 mr-1" /> Add Rule
                        </Button>
                    </div>

                    {(!currentPolicy.rules || currentPolicy.rules.length === 0) && (
                        <div className="text-sm text-muted-foreground italic p-2 border border-dashed rounded-md text-center">
                            No rules defined.
                        </div>
                    )}

                    {currentPolicy.rules?.map((rule, index) => (
                        <div key={index} className="flex gap-2 items-start">
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
