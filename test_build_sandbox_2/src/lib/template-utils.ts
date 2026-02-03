/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ServiceTemplate } from "@/lib/templates";
import { UpstreamServiceConfig } from "@/lib/client";

/**
 * Applies user-provided field values to a service template's configuration.
 *
 * @param template The service template.
 * @param fieldValues A map of field names to their values.
 * @returns A new UpstreamServiceConfig object with the values applied.
 */
export function applyTemplateFields(
  template: ServiceTemplate,
  fieldValues: Record<string, string>
): Partial<UpstreamServiceConfig> {
  // Deep clone the config to avoid mutating the original
  const config = JSON.parse(JSON.stringify(template.config));

  if (!template.fields) {
    return config;
  }

  for (const field of template.fields) {
    const value = fieldValues[field.name];
    if (value === undefined || value === "") {
        continue; // Skip empty values, or handle defaults if we had them
    }

    applyValueToConfig(config, field.key, value, field.replaceToken);
  }

  return config;
}

/**
 * Helper to set a value in a nested object based on a dot-notation path.
 * If replaceToken is provided, it performs a string replacement on the target value.
 */
function applyValueToConfig(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  obj: Record<string, any>,
  path: string,
  value: string,
  replaceToken?: string
) {
  const parts = path.split(".");
  let current = obj;

  for (let i = 0; i < parts.length - 1; i++) {
    const part = parts[i];
    if (!(part in current)) {
      // Create nested object if it doesn't exist
      current[part] = {};
    }
    current = current[part];
  }

  const lastPart = parts[parts.length - 1];

  if (replaceToken) {
    const currentValue = current[lastPart];
    if (typeof currentValue === "string") {
      current[lastPart] = currentValue.replace(replaceToken, value);
    } else {
        // Fallback: if target isn't a string, just set it (shouldn't happen for replaceToken use cases)
        current[lastPart] = value;
    }
  } else {
    // Direct assignment
    current[lastPart] = value;
  }
}
