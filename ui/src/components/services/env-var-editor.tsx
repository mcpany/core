/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Plus, X, Eye, EyeOff, Lock } from "lucide-react";

interface EnvVar {
  key: string;
  value: string;
  // If true, this variable is backed by a Secret ID (not plain text)
  isSecretRef: boolean;
  secretId?: string;
}

interface EnvVarEditorProps {
  initialEnv?: Record<string, { plainText?: string; secretId?: string }>;
  onChange: (env: Record<string, { plainText?: string; secretId?: string }>) => void;
}

export function EnvVarEditor({ initialEnv, onChange }: EnvVarEditorProps) {
  const [envVars, setEnvVars] = useState<EnvVar[]>(() => {
      if (!initialEnv) return [];
      return Object.entries(initialEnv).map(([key, val]) => {
          if (val.secretId) {
               return { key, value: val.secretId, isSecretRef: true, secretId: val.secretId };
          }
          return { key, value: val.plainText || "", isSecretRef: false };
      });
  });

  const [showValues, setShowValues] = useState<Record<number, boolean>>({});

  const updateParent = (vars: EnvVar[]) => {
      const newEnv: Record<string, { plainText?: string; secretId?: string }> = {};
      vars.forEach(v => {
          if (v.key) {
             if (v.isSecretRef && v.secretId) {
                 newEnv[v.key] = { secretId: v.secretId };
             } else {
                 newEnv[v.key] = { plainText: v.value };
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
              if (field === 'value' && v.isSecretRef) {
                  updated.isSecretRef = false;
                  updated.secretId = undefined;
              }
              return updated;
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
                  <Input
                      placeholder="KEY"
                      value={v.key}
                      onChange={(e) => updateVar(i, "key", e.target.value)}
                      className="flex-1"
                  />
                  <div className="relative flex-1">
                      {v.isSecretRef ? (
                           <div className="relative">
                               <Input
                                  value={`Secret: ${v.secretId}`}
                                  disabled
                                  className="pr-8 bg-muted text-muted-foreground"
                               />
                               <Lock className="absolute right-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                           </div>
                      ) : (
                          <>
                           <Input
                              placeholder="VALUE"
                              type={showValues[i] ? "text" : "password"}
                              value={v.value}
                              onChange={(e) => updateVar(i, "value", e.target.value)}
                              className="pr-8"
                          />
                           <button
                              type="button"
                              onClick={() => toggleShowValue(i)}
                              className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                           >
                              {showValues[i] ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                           </button>
                          </>
                      )}
                  </div>

                  <Button type="button" variant="ghost" size="icon" onClick={() => removeVar(i)}>
                      <X className="h-4 w-4" />
                  </Button>
              </div>
          ))}
      </div>
    </div>
  );
}
