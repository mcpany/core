/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * unwrapMcpResult serves as a public interface for interacting with unwrapMcpResult.
 *
 * Summary: Unwrap the mcp result appropriately based on current system conditions.
 *
 * Parameters:
 *   - Refer to the function signature for strongly-typed input arguments.
 *
 * Returns:
 *   - Returns the expected domain model or execution state.
 *
 * Throws/Errors:
 *   - Propagates exceptions from underlying validation layers.
 *
 * Side Effects:
 *   - May mutate state or perform network I/O depending on implementation.
 */
export function unwrapMcpResult(result: any): any {
    let content = result;

    // Unwrap CallToolResult structure
    if (result && typeof result === 'object' && Array.isArray(result.content)) {
        content = result.content;
    }

    // Handle Command Output wrapper
    if (content && typeof content === 'object' && !Array.isArray(content)) {
         if (content.stdout && typeof content.stdout === 'string') {
             try {
                 const inner = JSON.parse(content.stdout);
                 if (Array.isArray(inner) || (typeof inner === 'object' && inner !== null)) {
                     content = inner;
                 }
             } catch (e) {
                 // stdout is not JSON
             }
         }
    }

    // Handle deeply nested "content" (e.g. from stdout containing MCP content object)
    if (content && typeof content === 'object' && !Array.isArray(content) && Array.isArray(content.content)) {
        content = content.content;
    }

    // Additionally, if the content is an array of MCP Text items, parse the JSON inside text if possible
    if (Array.isArray(content)) {
        const isMcp = content.every((item: any) =>
            typeof item === 'object' && item !== null &&
            (item.type === 'text' || item.type === 'image' || item.type === 'resource')
        );

        if (isMcp) {
            // Only try to unwrap further if there's exactly one text block and it's JSON
            if (content.length === 1 && content[0].type === 'text' && typeof content[0].text === 'string') {
                try {
                    const parsed = JSON.parse(content[0].text);
                    if (typeof parsed === 'object' && parsed !== null) {
                        return unwrapMcpResult(parsed);
                    }
                } catch (e) {
                    // Not JSON inside text
                }
            }
            // For diffs, it's better to return the full array, but for SmartResultRenderer
            // we let the original mcpContent detection logic handle it.
        }
    }

    return content;
}

/**
 * deepParseJson serves as a public interface for interacting with deepParseJson.
 *
 * Summary: Deep the parse json appropriately based on current system conditions.
 *
 * Parameters:
 *   - Refer to the function signature for strongly-typed input arguments.
 *
 * Returns:
 *   - Returns the expected domain model or execution state.
 *
 * Throws/Errors:
 *   - Propagates exceptions from underlying validation layers.
 *
 * Side Effects:
 *   - May mutate state or perform network I/O depending on implementation.
 */
export function deepParseJson(obj: any): any {
    if (typeof obj === 'string') {
        try {
            const parsed = JSON.parse(obj);
            if (typeof parsed === 'object' && parsed !== null) {
                return deepParseJson(parsed);
            }
            // Only recurse for strings if the parsed string is still valid JSON
            // and it is different from the original string.
            // But if it's a primitive string (e.g. "just a string"), we shouldn't return parsed,
            // we should return obj to match previous behavior unless it's double escaped JSON.
            if (typeof parsed === 'string' && parsed !== obj) {
                 try {
                     const innerParsed = JSON.parse(parsed);
                     if (typeof innerParsed === 'object' && innerParsed !== null) {
                          return deepParseJson(parsed);
                     }
                 } catch (e) {}
            }
        } catch (e) {
            // Not a JSON string
        }
        return obj;
    }

    if (Array.isArray(obj)) {
        return obj.map(deepParseJson);
    }

    if (typeof obj === 'object' && obj !== null) {
        const result: any = {};
        for (const [key, value] of Object.entries(obj)) {
            result[key] = deepParseJson(value);
        }
        return result;
    }

    return obj;
}
