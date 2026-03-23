/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Unwraps an MCP Tool Result to extract its core content payload.
 * Useful for displaying clean data in tables or diffs without the protocol wrapper.
 *
 * @param result The raw tool execution result
 * @returns The unwrapped content or the original result if not a recognizable wrapper
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
                        return parsed;
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
 * Recursively traverses an object or array and parses any stringified JSON values.
 * This is particularly useful for diffing tool results where inner payloads might
 * be returned as strings within an MCP Text block, ensuring a rich diff view.
 *
 * @param obj The object or string to deeply parse
 * @returns The fully expanded object
 */
export function deepParseJson(obj: any): any {
    if (typeof obj === 'string') {
        try {
            const parsed = JSON.parse(obj);
            if (typeof parsed === 'object' && parsed !== null) {
                return deepParseJson(parsed);
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
