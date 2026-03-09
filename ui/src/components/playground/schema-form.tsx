/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";

interface SchemaFormProps {
  schema: Record<string, any>;
  value: Record<string, any>;
  onChange: (value: Record<string, any>) => void;
  /** Prefix for nested field paths */
  _prefix?: string;
  /** Parent value setter for nested objects */
  _parentSetter?: (key: string, val: any) => void;
  /** Parent key */
  _parentKey?: string;
}

/**
 * SchemaForm - renders a form based on a JSON Schema definition.
 *
 * @param props - Component props
 */
export function SchemaForm({
  schema,
  value,
  onChange,
  _prefix = "",
  _parentSetter,
  _parentKey,
}: SchemaFormProps) {
  if (!schema || schema.type !== "object" || !schema.properties) {
    return null;
  }

  const handleChange = (key: string, newVal: any) => {
    const updated = { ...value, [key]: newVal };
    if (_parentSetter && _parentKey !== undefined) {
      _parentSetter(_parentKey, updated);
    } else {
      onChange(updated);
    }
  };

  return (
    <div className="space-y-3">
      {Object.entries(schema.properties).map(([key, propSchema]: [string, any]) => {
        const fieldValue = value?.[key];
        const fieldLabel = key;

        if (propSchema.type === "object" && propSchema.properties) {
          return (
            <div key={key} className="space-y-2">
              <label className="block text-sm font-medium">{fieldLabel}</label>
              <div className="ml-4">
                <SchemaForm
                  schema={propSchema}
                  value={fieldValue || {}}
                  onChange={onChange}
                  _prefix={`${_prefix}${key}.`}
                  _parentSetter={handleChange}
                  _parentKey={key}
                />
              </div>
            </div>
          );
        }

        if (propSchema.type === "array") {
          const arrayValue: any[] = Array.isArray(fieldValue) ? fieldValue : [];
          return (
            <div key={key} className="space-y-2">
              <label className="block text-sm font-medium">{fieldLabel}</label>
              {arrayValue.map((_item, idx) => (
                <div key={idx} className="flex items-center gap-2">
                  <input
                    type="text"
                    className="border rounded px-2 py-1 text-sm"
                    placeholder={`Item ${idx + 1}`}
                    value={_item ?? ""}
                    onChange={(e) => {
                      const updated = [...arrayValue];
                      updated[idx] = e.target.value;
                      handleChange(key, updated);
                    }}
                  />
                </div>
              ))}
              <button
                type="button"
                className="text-sm text-blue-600"
                onClick={() => handleChange(key, [...arrayValue, undefined])}
              >
                Add Item
              </button>
            </div>
          );
        }

        if (propSchema.type === "boolean") {
          return (
            <div key={key} className="flex items-center gap-2">
              <input
                type="checkbox"
                id={`${_prefix}${key}`}
                checked={!!fieldValue}
                onChange={(e) => handleChange(key, e.target.checked)}
              />
              <label htmlFor={`${_prefix}${key}`} className="text-sm font-medium">
                {fieldLabel}
              </label>
            </div>
          );
        }

        const isNumeric = propSchema.type === "number" || propSchema.type === "integer";

        return (
          <div key={key} className="space-y-1">
            <label className="block text-sm font-medium">{fieldLabel}</label>
            <input
              type={isNumeric ? "number" : "text"}
              className="border rounded px-2 py-1 text-sm w-full"
              placeholder={isNumeric ? "0" : fieldLabel}
              value={fieldValue ?? ""}
              onChange={(e) => {
                const raw = e.target.value;
                const parsed = isNumeric ? Number(raw) : raw;
                handleChange(key, parsed);
              }}
            />
          </div>
        );
      })}
    </div>
  );
}
