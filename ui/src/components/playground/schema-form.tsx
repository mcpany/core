/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

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
      {Object.entries(schema.properties).map(
        ([key, propSchema]: [string, any]) => {
          const fieldValue = value?.[key];
          const fieldLabel = key;

          if (propSchema.type === "object" && propSchema.properties) {
            return (
              <div key={key} className="space-y-2">
                <label className="block text-sm font-medium">
                  {fieldLabel}
                </label>
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
            const arrayValue: any[] = Array.isArray(fieldValue)
              ? fieldValue
              : [];
            return (
              <div key={key} className="space-y-2">
                <label className="block text-sm font-medium">
                  {fieldLabel}
                </label>
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
                <label
                  htmlFor={`${_prefix}${key}`}
                  className="text-sm font-medium"
                >
                  {fieldLabel}
                </label>
              </div>
            );
          }

          if (
            propSchema.type === "string" &&
            (propSchema.contentEncoding === "base64" ||
              propSchema.format === "binary")
          ) {
            return (
              <div key={key} className="space-y-1">
                <label
                  htmlFor={`${_prefix}${key}`}
                  className="block text-sm font-medium"
                >
                  {fieldLabel}{" "}
                  <span className="text-xs text-muted-foreground">
                    (File Upload)
                  </span>
                </label>
                <input
                  id={`${_prefix}${key}`}
                  type="file"
                  className="border rounded px-2 py-1 text-sm w-full"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (!file) return;

                    const reader = new FileReader();
                    reader.onload = (ev) => {
                      const result = ev.target?.result as string;
                      // result is a data URL like "data:image/png;base64,iVBORw0KGgo..."
                      // We need to extract the base64 part
                      const base64Index = result.indexOf("base64,");
                      if (base64Index !== -1) {
                        const base64Data = result.substring(base64Index + 7);
                        handleChange(key, base64Data);
                      } else {
                        handleChange(key, result);
                      }
                    };
                    reader.readAsDataURL(file);
                  }}
                />
                {fieldValue && (
                  <div className="text-xs text-green-600 truncate mt-1">
                    File loaded (
                    {fieldValue.length > 50
                      ? fieldValue.substring(0, 50) + "..."
                      : fieldValue}
                    )
                  </div>
                )}
              </div>
            );
          }

          const isNumeric =
            propSchema.type === "number" || propSchema.type === "integer";

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
        },
      )}
    </div>
  );
}
