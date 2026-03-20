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

        // Check if there are other properties besides content and isError
        // If there are only standard wrapper properties (like isError), we consider it fully wrappable.
        // If it has non-standard properties, we might still want to unwrap content, but the caller
        // (like diff viewer) needs to see the full structure if it has extra data.
        const keys = Object.keys(result);
        const hasExtraKeys = keys.some(k => k !== 'content' && k !== 'isError');

        if (hasExtraKeys) {
            // If it has extra keys (like "value" in the diff-feature.png), it means the result IS NOT
            // just a simple wrapper. So we should NOT strip the content wrapper. We should return the
            // original result, but let deepParseJson handle deep stringified JSONs.
            // However, we want deepParseJson to parse it.
            // deepParseJson handles nested objects, so if we return result, deepParseJson WILL parse it.
            return result;
        }
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

    // If we originally found a clean CallToolResult but couldn't unwrap its inner array to a single JSON object,
    // we should still return the content array (which is what the old logic did) so it gets unwrapped.
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
