/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Plus, X, Eye, EyeOff, Lock, Unlock, Key, Sparkles } from "lucide-react";
import { SecretPicker } from "@/components/secrets/secret-picker";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

interface EnvVar {
  key: string;
  value: string;
  // If true, this variable is backed by a Secret ID (not plain text)
  isSecretRef: boolean;
  secretId?: string;
  isSuggested?: boolean;
}

import { SecretValue } from "@proto/config/v1/auth";

interface EnvVarEditorProps {
  initialEnv?: Record<string, SecretValue>;
  suggestedKeys?: Record<string, string>; // Key -> Description
  onChange: (env: Record<string, SecretValue>) => void;
}

/**
 * EnvVarEditor.
 *
 * @param onChange - The onChange.
 */
export function EnvVarEditor({ initialEnv, suggestedKeys, onChange }: EnvVarEditorProps) {
  const [envVars, setEnvVars] = useState<EnvVar[]>(() => {
      const vars: EnvVar[] = [];
      const usedKeys = new Set<string>();

      // Load initial env
      if (initialEnv) {
          Object.entries(initialEnv).forEach(([key, val]) => {
              usedKeys.add(key);
              if (val.secretId) {
                  vars.push({ key, value: val.secretId, isSecretRef: true, secretId: val.secretId });
              } else {
                  // Heuristic: if plainText starts with ${ and ends with }, treat as secret ref
                  const plainText = val.plainText || "";
                  const secretMatch = plainText.match(/^\$\{(.+)\}$/);
                  if (secretMatch) {
                       vars.push({ key, value: secretMatch[1], isSecretRef: true, secretId: secretMatch[1] });
                  } else {
                       vars.push({ key, value: plainText, isSecretRef: false });
                  }
              }
          });
      }

      // Merge suggestions
      if (suggestedKeys) {
          Object.keys(suggestedKeys).forEach(key => {
              if (!usedKeys.has(key)) {
                  vars.push({
                      key,
                      value: "",
                      isSecretRef: false,
                      isSuggested: true
                  });
              }
          });
      }

      return vars;
  });

  const [showValues, setShowValues] = useState<Record<number, boolean>>({});

  const updateParent = (vars: EnvVar[]) => {
      const newEnv: Record<string, SecretValue> = {};
      vars.forEach(v => {
          if (v.key) {
             if (v.isSecretRef && v.secretId) {
                 newEnv[v.key] = { plainText: `\${${v.secretId}}`, validationRegex: "" };
             } else {
                 newEnv[v.key] = { plainText: v.value, validationRegex: "" };
             }
          }
      });
      onChange(newEnv);
  };

  const addVar = () => {
      setEnvVars([...envVars, { key: "", value: "", isSecretRef: false }]);
  };

  const removeVar = (index: number) => {
      const newVars = envVars.filter((_, i) => i !== index);
      setEnvVars(newVars);
      updateParent(newVars);
  };

  const updateVar = (index: number, field: keyof EnvVar, value: string | boolean) => {
      const newVars = envVars.map((v, i) => {
          if (i === index) {
              const updated = { ...v, [field]: value };
              // If user edits value of a secret ref, it becomes plain text unless we implement secret picker
              if (field === 'secretId') {
                  updated.secretId = value as string;
              }
              // If user edits key, it's no longer a suggested entry tied to the original suggestion logic
              // but we keep isSuggested true until they save? No, let's clear it if they change the key.
              if (field === 'key' && v.isSuggested && value !== v.key) {
                  updated.isSuggested = false;
              }
              return updated;
          }
          return v;
      });
      setEnvVars(newVars);
      updateParent(newVars);
  };

  const toggleSecretMode = (index: number) => {
      const newVars = envVars.map((v, i) => {
          if (i === index) {
              const newIsSecretRef = !v.isSecretRef;
              return {
                  ...v,
                  isSecretRef: newIsSecretRef,
                  // If switching to secret, clear value. If to text, clear secretId.
                  value: newIsSecretRef ? "" : "",
                  secretId: newIsSecretRef ? "" : undefined
              };
          }
          return v;
      });
      setEnvVars(newVars);
      updateParent(newVars);
  };

  const toggleShowValue = (index: number) => {
      setShowValues(prev => ({ ...prev, [index]: !prev[index] }));
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
          <Label>Environment Variables</Label>
          <Button type="button" variant="outline" size="sm" onClick={addVar}>
              <Plus className="mr-2 h-3 w-3" /> Add Variable
          </Button>
      </div>

      {envVars.length === 0 && (
          <div className="text-sm text-muted-foreground italic border border-dashed rounded p-4 text-center">
              No environment variables set.
          </div>
      )}

      <div className="space-y-2">
          {envVars.map((v, i) => (
              <div key={i} className="flex items-center gap-2">
                   <div className="relative flex-1">
                      <Input
                          placeholder="KEY"
                          value={v.key}
                          onChange={(e) => updateVar(i, "key", e.target.value)}
                          className={v.isSuggested ? "border-blue-300 dark:border-blue-700 bg-blue-50/10" : ""}
                      />
                      {v.isSuggested && suggestedKeys && suggestedKeys[v.key] && (
                           <TooltipProvider>
                                <Tooltip>
                                    <TooltipTrigger asChild>
                                        <div className="absolute right-2 top-1/2 -translate-y-1/2 cursor-help">
                                            <Badge variant="secondary" className="text-[10px] h-4 px-1 flex items-center gap-1">
                                                <Sparkles className="h-2 w-2 text-primary" /> Suggested
                                            </Badge>
                                        </div>
                                    </TooltipTrigger>
                                    <TooltipContent>
                                        <p>{suggestedKeys[v.key]}</p>
                                    </TooltipContent>
                                </Tooltip>
                           </TooltipProvider>
                      )}
                   </div>

                  <div className="relative flex-1">
                      {v.isSecretRef ? (
                           <SecretPicker
                                value={v.secretId}
                                onSelect={(key) => updateVar(i, "secretId", key)}
                           >
                                <div className="relative cursor-pointer group">
                                    <Input
                                        value={v.secretId || ""}
                                        readOnly
                                        className="pr-8 bg-muted/50 cursor-pointer text-primary font-medium focus-visible:ring-primary"
                                        placeholder={suggestedKeys?.[v.key] ? "Set secret..." : "Select a secret..."}
                                    />
                                    <Key className="absolute right-2 top-1/2 -translate-y-1/2 h-4 w-4 text-primary group-hover:scale-110 transition-transform" />
                                </div>
                           </SecretPicker>
                      ) : (
                          <>
                           <Input
                              placeholder={suggestedKeys?.[v.key] || "VALUE"}
                              type={showValues[i] ? "text" : "password"}
                              value={v.value}
                              onChange={(e) => updateVar(i, "value", e.target.value)}
                              className="pr-10"
                          />
                           <button
                              type="button"
                              onClick={() => toggleShowValue(i)}
                              className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground p-1"
                           >
                              {showValues[i] ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                           </button>
                          </>
                      )}
                  </div>

                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => toggleSecretMode(i)}
                    className={v.isSecretRef ? "text-primary hover:text-primary/80" : "text-muted-foreground"}
                    title={v.isSecretRef ? "Switch to Plain Text" : "Switch to Secret Reference"}
                  >
                      {v.isSecretRef ? <Lock className="h-4 w-4" /> : <Unlock className="h-4 w-4" />}
                  </Button>

                  <Button type="button" variant="ghost" size="icon" onClick={() => removeVar(i)} className="text-destructive/50 hover:text-destructive">
                      <X className="h-4 w-4" />
                  </Button>
              </div>
          ))}
      </div>
    </div>
  );
}
