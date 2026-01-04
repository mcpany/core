/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useForm, Controller } from "react-hook-form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { ToolDefinition } from "@/lib/client";
import { Info } from "lucide-react";

interface ToolFormProps {
  tool: ToolDefinition;
  onSubmit: (data: any) => void;
  onCancel: () => void;
}

export function ToolForm({ tool, onSubmit, onCancel }: ToolFormProps) {
  const { register, control, handleSubmit, formState: { errors } } = useForm();
  const schema = tool.schema || {};
  const properties = schema.properties || {};
  const required = schema.required || [];

  const handleFormSubmit = (data: any) => {
    // Post-process data to match types
    const processedData: any = {};
    Object.keys(data).forEach((key) => {
      const prop = properties[key];
      const value = data[key];

      if (value === "" || value === undefined) return; // Skip empty optional fields

      if (prop.type === "integer" || prop.type === "number") {
        processedData[key] = Number(value);
      } else if (prop.type === "boolean") {
        processedData[key] = Boolean(value);
      } else {
        processedData[key] = value;
      }
    });
    onSubmit(processedData);
  };

  return (
    <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-4 py-2">
      <div className="space-y-4">
        {Object.keys(properties).length === 0 && (
           <div className="text-sm text-muted-foreground italic">
               This tool takes no arguments.
           </div>
        )}
        {Object.keys(properties).map((key) => {
          const prop = properties[key];
          const isRequired = required.includes(key);

          return (
            <div key={key} className="space-y-2">
              <Label htmlFor={key} className="flex items-center gap-1">
                {key}
                {isRequired && <span className="text-red-500">*</span>}
                <span className="text-xs font-normal text-muted-foreground ml-2">({prop.type})</span>
              </Label>

              {prop.description && (
                  <p className="text-[10px] text-muted-foreground">{prop.description}</p>
              )}

              {prop.enum ? (
                <Controller
                  name={key}
                  control={control}
                  rules={{ required: isRequired }}
                  render={({ field }) => (
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <SelectTrigger>
                        <SelectValue placeholder={`Select ${key}`} />
                      </SelectTrigger>
                      <SelectContent>
                        {prop.enum.map((val: string) => (
                          <SelectItem key={val} value={val}>
                            {val}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
              ) : prop.type === "boolean" ? (
                 <Controller
                  name={key}
                  control={control}
                  render={({ field }) => (
                    <div className="flex items-center gap-2">
                        <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                        />
                        <span className="text-sm text-muted-foreground">{field.value ? "True" : "False"}</span>
                    </div>
                  )}
                />
              ) : prop.type === "integer" || prop.type === "number" ? (
                <Input
                  id={key}
                  type="number"
                  step={prop.type === "integer" ? "1" : "any"}
                  placeholder={`Enter ${key}`}
                  {...register(key, { required: isRequired })}
                />
              ) : (
                 <Input // Default to text for string and others
                  id={key}
                  placeholder={`Enter ${key}`}
                  {...register(key, { required: isRequired })}
                />
              )}

              {errors[key] && (
                <span className="text-xs text-red-500">This field is required</span>
              )}
            </div>
          );
        })}
      </div>

      <div className="flex justify-end gap-2 pt-4 border-t mt-4">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit">
          Run Tool
        </Button>
      </div>
    </form>
  );
}
