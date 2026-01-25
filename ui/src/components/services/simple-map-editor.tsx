/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Plus, X } from "lucide-react";

interface SimpleMapEditorProps {
  initialData?: Record<string, string>;
  onChange: (data: Record<string, string>) => void;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
  label?: string;
}

export function SimpleMapEditor({ initialData, onChange, keyPlaceholder = "Key", valuePlaceholder = "Value", label = "Items" }: SimpleMapEditorProps) {
  const [items, setItems] = useState<{ key: string, value: string }[]>(() => {
      if (!initialData) return [];
      return Object.entries(initialData).map(([key, value]) => ({ key, value }));
  });

  const updateParent = (currentItems: { key: string, value: string }[]) => {
      const newData: Record<string, string> = {};
      currentItems.forEach(item => {
          if (item.key) {
              newData[item.key] = item.value;
          }
      });
      onChange(newData);
  };

  const addItem = () => {
      setItems([...items, { key: "", value: "" }]);
  };

  const removeItem = (index: number) => {
      const newItems = items.filter((_, i) => i !== index);
      setItems(newItems);
      updateParent(newItems);
  };

  const updateItem = (index: number, field: 'key' | 'value', val: string) => {
      const newItems = items.map((item, i) => {
          if (i === index) {
              return { ...item, [field]: val };
          }
          return item;
      });
      setItems(newItems);
      updateParent(newItems);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
          <Label>{label}</Label>
          <Button type="button" variant="outline" size="sm" onClick={addItem}>
              <Plus className="mr-2 h-3 w-3" /> Add Item
          </Button>
      </div>

      {items.length === 0 && (
          <div className="text-sm text-muted-foreground italic border border-dashed rounded p-4 text-center">
              No items configured.
          </div>
      )}

      <div className="space-y-2">
          {items.map((item, i) => (
              <div key={i} className="flex items-center gap-2">
                  <Input
                      placeholder={keyPlaceholder}
                      value={item.key}
                      onChange={(e) => updateItem(i, "key", e.target.value)}
                      className="flex-1"
                  />
                  <Input
                      placeholder={valuePlaceholder}
                      value={item.value}
                      onChange={(e) => updateItem(i, "value", e.target.value)}
                      className="flex-1"
                  />
                  <Button type="button" variant="ghost" size="icon" onClick={() => removeItem(i)}>
                      <X className="h-4 w-4" />
                  </Button>
              </div>
          ))}
      </div>
    </div>
  );
}
