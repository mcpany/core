/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from "react";
import { ServiceTemplate } from "@/lib/templates";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

interface TemplateConfigFormProps {
  template: ServiceTemplate;
  onCancel: () => void;
  onSubmit: (values: Record<string, string>) => void;
}

export function TemplateConfigForm({ template, onCancel, onSubmit }: TemplateConfigFormProps) {
  const [values, setValues] = useState<Record<string, string>>({});

  const handleChange = (name: string, value: string) => {
    setValues((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(values);
  };

  return (
    <div className="space-y-6">
      <CardHeader className="px-0">
        <div className="flex items-center gap-2">
            <div className="p-2 bg-primary/10 rounded-md">
                <template.icon className="w-6 h-6 text-primary" />
            </div>
            <div>
                <CardTitle>{template.name}</CardTitle>
                <CardDescription>{template.description}</CardDescription>
            </div>
        </div>
      </CardHeader>

      <form onSubmit={handleSubmit}>
        <div className="space-y-4">
          {template.fields?.map((field) => (
            <div key={field.name} className="space-y-2">
              <Label htmlFor={field.name}>{field.label}</Label>
              <Input
                id={field.name}
                placeholder={field.placeholder}
                value={values[field.name] || ""}
                onChange={(e) => handleChange(field.name, e.target.value)}
                required
              />
              {field.replaceToken && (
                  <p className="text-xs text-muted-foreground">
                      This will replace <code>{field.replaceToken}</code> in the configuration.
                  </p>
              )}
            </div>
          ))}
        </div>

        <div className="flex justify-end gap-2 mt-8">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit">
            Continue
          </Button>
        </div>
      </form>
    </div>
  );
}
