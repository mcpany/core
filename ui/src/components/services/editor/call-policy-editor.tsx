/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Trash2, Plus, GripVertical, ShieldCheck, ShieldAlert, Shield, Play } from "lucide-react";
import { CallPolicy, CallPolicy_Action, CallPolicyRule } from "@proto/config/v1/upstream_service";
import { DragDropContext, Droppable, Draggable, DropResult } from "@hello-pangea/dnd";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";

interface CallPolicyEditorProps {
    policy: CallPolicy | undefined;
    onChange: (policy: CallPolicy) => void;
}

/**
 * A component for editing runtime call policies (Firewall).
 * Allows setting a default action and adding specific rules based on regex for tool names and arguments.
 * Includes a simulator to test rules.
 *
 * @param props - The component props.
 * @param props.policy - The current policy object.
 * @param props.onChange - Callback invoked when the policy is modified.
 * @returns The rendered policy editor component.
 */
export function CallPolicyEditor({ policy, onChange }: CallPolicyEditorProps) {
    // Default to empty policy if undefined
    const currentPolicy: CallPolicy = policy || {
        defaultAction: CallPolicy_Action.ALLOW,
        rules: []
    };

    // Simulator State
    const [simToolName, setSimToolName] = useState("delete_user");
    const [simArgs, setSimArgs] = useState('{"id": 123}');
    const [simResult, setSimResult] = useState<{ action: CallPolicy_Action, ruleIndex: number } | null>(null);

    const handleDefaultActionChange = (val: string) => {
        onChange({
            ...currentPolicy,
            defaultAction: parseInt(val) as CallPolicy_Action
        });
    };

    const addRule = () => {
        const newRules = [
            ...(currentPolicy.rules || []),
            {
                action: CallPolicy_Action.DENY,
                nameRegex: "",
                argumentRegex: "",
                urlRegex: "",
                callIdRegex: ""
            }
        ];
        onChange({ ...currentPolicy, rules: newRules });
    };

    const updateRule = (index: number, updates: Partial<CallPolicyRule>) => {
        const newRules = [...(currentPolicy.rules || [])];
        newRules[index] = { ...newRules[index], ...updates };
        onChange({ ...currentPolicy, rules: newRules });
    };

    const removeRule = (index: number) => {
        const newRules = [...(currentPolicy.rules || [])];
        newRules.splice(index, 1);
        onChange({ ...currentPolicy, rules: newRules });
    };

    const onDragEnd = (result: DropResult) => {
        if (!result.destination) return;
        const newRules = Array.from(currentPolicy.rules || []);
        const [reorderedItem] = newRules.splice(result.source.index, 1);
        newRules.splice(result.destination.index, 0, reorderedItem);
        onChange({ ...currentPolicy, rules: newRules });
    };

    // Simulator Logic
    const runSimulation = () => {
        let matchedRuleIndex = -1;
        let finalAction = currentPolicy.defaultAction;

        const rules = currentPolicy.rules || [];
        for (let i = 0; i < rules.length; i++) {
            const rule = rules[i];
            let matches = true;

            if (rule.nameRegex) {
                try {
                    const regex = new RegExp(rule.nameRegex);
                    if (!regex.test(simToolName)) matches = false;
                } catch { matches = false; }
            }

            if (matches && rule.argumentRegex) {
                try {
                    const regex = new RegExp(rule.argumentRegex);
                    if (!regex.test(simArgs)) matches = false;
                } catch { matches = false; }
            }

            if (matches) {
                finalAction = rule.action;
                matchedRuleIndex = i;
                break;
            }
        }

        setSimResult({ action: finalAction, ruleIndex: matchedRuleIndex });
    };

    const getActionLabel = (action: CallPolicy_Action) => {
        switch (action) {
            case CallPolicy_Action.ALLOW: return { label: "Allow", color: "bg-green-500" };
            case CallPolicy_Action.DENY: return { label: "Deny", color: "bg-red-500" };
            case CallPolicy_Action.SAVE_CACHE: return { label: "Cache", color: "bg-blue-500" };
            case CallPolicy_Action.DELETE_CACHE: return { label: "Clear Cache", color: "bg-amber-500" };
            default: return { label: "Unknown", color: "bg-gray-500" };
        }
    };

    return (
        <Card>
            <CardHeader>
                <div className="flex items-center justify-between">
                    <div>
                        <CardTitle className="text-base flex items-center gap-2">
                            <Shield className="h-5 w-5 text-primary" />
                            Security Firewall (Call Policies)
                        </CardTitle>
                        <CardDescription>
                            Define runtime rules to allow or deny tool executions based on name and arguments.
                        </CardDescription>
                    </div>
                    <div className="flex items-center gap-2">
                        <Label className="text-sm font-medium">Default Action:</Label>
                        <Select
                            value={currentPolicy.defaultAction.toString()}
                            onValueChange={handleDefaultActionChange}
                        >
                            <SelectTrigger className="w-[140px] h-8">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value={CallPolicy_Action.ALLOW.toString()}>Allow All</SelectItem>
                                <SelectItem value={CallPolicy_Action.DENY.toString()}>Deny All</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>
            </CardHeader>
            <CardContent className="space-y-6">
                <DragDropContext onDragEnd={onDragEnd}>
                    <Droppable droppableId="firewall-rules">
                        {(provided) => (
                            <div
                                {...provided.droppableProps}
                                ref={provided.innerRef}
                                className="space-y-2"
                            >
                                <div className="flex items-center justify-between mb-2">
                                    <Label className="text-xs font-semibold uppercase text-muted-foreground">Firewall Rules (First Match Wins)</Label>
                                    <Button variant="ghost" size="sm" onClick={addRule} className="h-6 text-xs gap-1">
                                        <Plus className="h-3 w-3" /> Add Rule
                                    </Button>
                                </div>

                                {(!currentPolicy.rules || currentPolicy.rules.length === 0) && (
                                    <div className="text-sm text-muted-foreground italic p-4 border border-dashed rounded-md text-center bg-muted/20">
                                        No active rules. Traffic matches Default Action.
                                    </div>
                                )}

                                {currentPolicy.rules?.map((rule, index) => (
                                    <Draggable key={`rule-${index}`} draggableId={`rule-${index}`} index={index}>
                                        {(provided) => (
                                            <div
                                                ref={provided.innerRef}
                                                {...provided.draggableProps}
                                                className="flex items-start gap-3 p-3 bg-muted/10 border rounded-md group"
                                            >
                                                <div {...provided.dragHandleProps} className="mt-2 cursor-grab text-muted-foreground hover:text-foreground">
                                                    <GripVertical className="h-4 w-4" />
                                                </div>

                                                <div className="flex-1 grid grid-cols-1 md:grid-cols-12 gap-3">
                                                    <div className="md:col-span-2">
                                                        <Label className="text-[10px] uppercase text-muted-foreground">Action</Label>
                                                        <Select
                                                            value={rule.action.toString()}
                                                            onValueChange={(val) => updateRule(index, { action: parseInt(val) as CallPolicy_Action })}
                                                        >
                                                            <SelectTrigger className="h-8">
                                                                <SelectValue />
                                                            </SelectTrigger>
                                                            <SelectContent>
                                                                <SelectItem value={CallPolicy_Action.ALLOW.toString()}>Allow</SelectItem>
                                                                <SelectItem value={CallPolicy_Action.DENY.toString()}>Deny</SelectItem>
                                                                <SelectItem value={CallPolicy_Action.SAVE_CACHE.toString()}>Cache</SelectItem>
                                                                <SelectItem value={CallPolicy_Action.DELETE_CACHE.toString()}>Clear Cache</SelectItem>
                                                            </SelectContent>
                                                        </Select>
                                                    </div>

                                                    <div className="md:col-span-4">
                                                        <Label className="text-[10px] uppercase text-muted-foreground">Tool Name (Regex)</Label>
                                                        <Input
                                                            value={rule.nameRegex}
                                                            onChange={(e) => updateRule(index, { nameRegex: e.target.value })}
                                                            className="h-8 font-mono text-xs"
                                                            placeholder=".*"
                                                        />
                                                    </div>

                                                    <div className="md:col-span-6">
                                                        <Label className="text-[10px] uppercase text-muted-foreground">Arguments (JSON Regex)</Label>
                                                        <Input
                                                            value={rule.argumentRegex}
                                                            onChange={(e) => updateRule(index, { argumentRegex: e.target.value })}
                                                            className="h-8 font-mono text-xs"
                                                            placeholder=".*"
                                                        />
                                                    </div>
                                                </div>

                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    onClick={() => removeRule(index)}
                                                    className="h-8 w-8 text-muted-foreground hover:text-destructive mt-4"
                                                >
                                                    <Trash2 className="h-4 w-4" />
                                                </Button>
                                            </div>
                                        )}
                                    </Draggable>
                                ))}
                                {provided.placeholder}
                            </div>
                        )}
                    </Droppable>
                </DragDropContext>

                <Separator />

                <div className="space-y-4">
                    <Label className="text-sm font-semibold flex items-center gap-2">
                        <Play className="h-4 w-4" /> Policy Simulator
                    </Label>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <Label htmlFor="sim-tool" className="text-xs">Tool Name</Label>
                            <Input
                                id="sim-tool"
                                value={simToolName}
                                onChange={(e) => setSimToolName(e.target.value)}
                                className="font-mono text-xs"
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="sim-args" className="text-xs">Arguments (JSON String)</Label>
                            <Input
                                id="sim-args"
                                value={simArgs}
                                onChange={(e) => setSimArgs(e.target.value)}
                                className="font-mono text-xs"
                            />
                        </div>
                    </div>
                    <div className="flex items-center justify-between bg-muted/30 p-3 rounded-md border">
                        <div className="flex items-center gap-4">
                            <Button size="sm" onClick={runSimulation} variant="outline">Test Rules</Button>
                            {simResult && (
                                <div className="flex items-center gap-2 animate-in fade-in zoom-in-95">
                                    <span className="text-sm text-muted-foreground">Result:</span>
                                    <Badge className={getActionLabel(simResult.action).color}>
                                        {getActionLabel(simResult.action).label}
                                    </Badge>
                                    <span className="text-xs text-muted-foreground">
                                        {simResult.ruleIndex === -1 ? "(Default Policy)" : `(Matched Rule #${simResult.ruleIndex + 1})`}
                                    </span>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
