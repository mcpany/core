/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { CallPolicy, CallPolicyRule, CallPolicy_Action } from "@proto/config/v1/upstream_service";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Trash2, Shield, GripVertical } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

interface CallPolicyEditorProps {
  policies: CallPolicy[];
  onChange: (policies: CallPolicy[]) => void;
}

const ACTION_LABELS: Record<number, string> = {
  [CallPolicy_Action.ALLOW]: "Allow",
  [CallPolicy_Action.DENY]: "Deny",
  [CallPolicy_Action.SAVE_CACHE]: "Save Cache",
  [CallPolicy_Action.DELETE_CACHE]: "Delete Cache",
};

export function CallPolicyEditor({ policies, onChange }: CallPolicyEditorProps) {
  const addPolicy = () => {
    const newPolicy: CallPolicy = {
      defaultAction: CallPolicy_Action.DENY, // Default to secure
      rules: [],
    };
    onChange([...policies, newPolicy]);
  };

  const removePolicy = (index: number) => {
    const newPolicies = [...policies];
    newPolicies.splice(index, 1);
    onChange(newPolicies);
  };

  const updatePolicy = (index: number, updated: CallPolicy) => {
    const newPolicies = [...policies];
    newPolicies[index] = updated;
    onChange(newPolicies);
  };

  const addRule = (policyIndex: number) => {
    const policy = policies[policyIndex];
    const newRule: CallPolicyRule = {
      action: CallPolicy_Action.ALLOW,
      nameRegex: "",
      argumentRegex: "",
      urlRegex: "",
      callIdRegex: "",
    };
    const updatedPolicy = {
      ...policy,
      rules: [...policy.rules, newRule],
    };
    updatePolicy(policyIndex, updatedPolicy);
  };

  const removeRule = (policyIndex: number, ruleIndex: number) => {
    const policy = policies[policyIndex];
    const newRules = [...policy.rules];
    newRules.splice(ruleIndex, 1);
    updatePolicy(policyIndex, { ...policy, rules: newRules });
  };

  const updateRule = (policyIndex: number, ruleIndex: number, updatedRule: CallPolicyRule) => {
    const policy = policies[policyIndex];
    const newRules = [...policy.rules];
    newRules[ruleIndex] = updatedRule;
    updatePolicy(policyIndex, { ...policy, rules: newRules });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="space-y-1">
            <h3 className="text-lg font-medium">Access Control Policies</h3>
            <p className="text-sm text-muted-foreground">
                Define rules to allow or deny specific tool executions. Policies are evaluated in order.
            </p>
        </div>
        <Button onClick={addPolicy} variant="outline" size="sm">
          <Plus className="mr-2 h-4 w-4" /> Add Policy
        </Button>
      </div>

      {policies.length === 0 && (
        <div className="flex flex-col items-center justify-center p-8 border-2 border-dashed rounded-lg bg-muted/10 text-muted-foreground">
          <Shield className="h-10 w-10 mb-2 opacity-20" />
          <p>No policies defined.</p>
          <p className="text-xs">All calls will be allowed by default unless restricted.</p>
        </div>
      )}

      {policies.map((policy, policyIndex) => (
        <Card key={policyIndex} className="relative group border-l-4 border-l-primary">
          <Button
            variant="ghost"
            size="icon"
            className="absolute top-2 right-2 text-muted-foreground hover:text-destructive"
            onClick={() => removePolicy(policyIndex)}
          >
            <Trash2 className="h-4 w-4" />
          </Button>

          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
                Policy #{policyIndex + 1}
            </CardTitle>
            <div className="flex items-center gap-4 mt-2">
                <div className="flex items-center gap-2">
                    <Label htmlFor={`default-action-${policyIndex}`} className="text-sm whitespace-nowrap">Default Action:</Label>
                    <Select
                        value={String(policy.defaultAction)}
                        onValueChange={(v) => updatePolicy(policyIndex, { ...policy, defaultAction: parseInt(v) })}
                    >
                        <SelectTrigger className="w-[140px] h-8">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value={String(CallPolicy_Action.ALLOW)}>Allow</SelectItem>
                            <SelectItem value={String(CallPolicy_Action.DENY)}>Deny</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            </div>
          </CardHeader>
          <Separator />
          <CardContent className="pt-4 space-y-4">
            <div className="space-y-2">
                <div className="flex items-center justify-between">
                    <Label className="text-xs font-semibold uppercase text-muted-foreground">Rules</Label>
                    <Button onClick={() => addRule(policyIndex)} variant="secondary" size="sm" className="h-7 text-xs">
                        <Plus className="mr-1 h-3 w-3" /> Add Rule
                    </Button>
                </div>

                {policy.rules.length === 0 ? (
                    <div className="text-xs text-muted-foreground italic py-2">
                        No specific rules. Default action applies to everything.
                    </div>
                ) : (
                    <div className="space-y-3">
                        {policy.rules.map((rule, ruleIndex) => (
                            <div key={ruleIndex} className="flex gap-3 items-start p-3 bg-muted/40 rounded-md border text-sm">
                                <div className="mt-2 text-muted-foreground">
                                    <span className="text-xs font-mono">{ruleIndex + 1}.</span>
                                </div>
                                <div className="grid grid-cols-1 md:grid-cols-4 gap-3 flex-1">
                                    <div className="space-y-1">
                                        <Label className="text-[10px] text-muted-foreground">Action</Label>
                                        <Select
                                            value={String(rule.action)}
                                            onValueChange={(v) => updateRule(policyIndex, ruleIndex, { ...rule, action: parseInt(v) })}
                                        >
                                            <SelectTrigger className="h-8">
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {Object.entries(ACTION_LABELS).map(([val, label]) => (
                                                    <SelectItem key={val} value={val}>{label}</SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                    </div>
                                    <div className="space-y-1 md:col-span-3 grid grid-cols-1 md:grid-cols-3 gap-2">
                                        <div className="space-y-1">
                                            <Label className="text-[10px] text-muted-foreground">Tool Name (Regex)</Label>
                                            <Input
                                                value={rule.nameRegex || ""}
                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { ...rule, nameRegex: e.target.value })}
                                                placeholder="e.g. ^list_.*"
                                                className="h-8 font-mono text-xs"
                                            />
                                        </div>
                                        <div className="space-y-1">
                                            <Label className="text-[10px] text-muted-foreground">Args (JSON Regex)</Label>
                                            <Input
                                                value={rule.argumentRegex || ""}
                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { ...rule, argumentRegex: e.target.value })}
                                                placeholder='e.g. "force": true'
                                                className="h-8 font-mono text-xs"
                                            />
                                        </div>
                                        <div className="space-y-1">
                                            <Label className="text-[10px] text-muted-foreground">URL/Path (Regex)</Label>
                                            <Input
                                                value={rule.urlRegex || ""}
                                                onChange={(e) => updateRule(policyIndex, ruleIndex, { ...rule, urlRegex: e.target.value })}
                                                placeholder="e.g. /v1/dangerous"
                                                className="h-8 font-mono text-xs"
                                            />
                                        </div>
                                    </div>
                                </div>
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8 text-muted-foreground hover:text-destructive mt-4"
                                    onClick={() => removeRule(policyIndex, ruleIndex)}
                                >
                                    <Trash2 className="h-4 w-4" />
                                </Button>
                            </div>
                        ))}
                    </div>
                )}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
