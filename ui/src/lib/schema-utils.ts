/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Recursively finds the path to the first field in the schema that accepts binary data.
 * Binary data is identified by `contentEncoding: "base64"` or `format: "binary"`.
 *
 * @param schema The JSON schema to search.
 * @param currentPath The current path prefix (used for recursion).
 * @returns The dot-notation path to the field (e.g., "image.data"), or null if not found.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function findFileFieldPath(schema: any, currentPath = ""): string | null {
  if (!schema || typeof schema !== "object") {
    return null;
  }

  // Check current node
  if (schema.contentEncoding === "base64" || schema.format === "binary") {
    return currentPath;
  }

  // Check properties if object
  if (schema.type === "object" && schema.properties) {
    for (const key in schema.properties) {
      const propPath = currentPath ? `${currentPath}.${key}` : key;
      const found = findFileFieldPath(schema.properties[key], propPath);
      if (found) {
        return found;
      }
    }
  }

  return null;
}

/**
 * Sets a value deeply in an object using a dot-notation path.
 * Creates nested objects as needed.
 *
 * @param obj The object to modify.
 * @param path The dot-notation path (e.g. "a.b.c").
 * @param value The value to set.
 * @returns A new object with the value set (immutabilityish - shallow copy of modified paths).
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function setDeepValue(obj: any, path: string, value: any): any {
  if (!path) return value;

  const keys = path.split('.');
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const result = { ...obj };
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let current = result;

  for (let i = 0; i < keys.length; i++) {
    const key = keys[i];
    if (i === keys.length - 1) {
      current[key] = value;
    } else {
      current[key] = { ...current[key] } || {};
      current = current[key];
    }
  }

  return result;
}
