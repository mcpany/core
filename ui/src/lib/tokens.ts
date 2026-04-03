/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * estimateTokens serves as a public interface for interacting with estimateTokens.
 *
 * Summary: Estimate the tokens appropriately based on current system conditions.
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
export function estimateTokens(input: any): number {
    if (!input) return 0;

    const text = typeof input === 'string' ? input : JSON.stringify(input);

    // Simple heuristic used by many LLM providers for estimation:
    // Approximately 4 characters per token for English text.
    // We add some overhead for whitespace and special characters.
    const charCount = text.length;
    const wordCount = text.trim().split(/\s+/).length;

    // Heuristic 1: 4 chars per token
    // Heuristic 2: 1.3 words per token
    // We'll take a balanced approach or the max of both for safety.
    const h1 = Math.ceil(charCount / 4);
    const h2 = Math.ceil(wordCount * 1.3);

    return Math.max(h1, h2);
}

/**
 * estimateMessageTokens serves as a public interface for interacting with estimateMessageTokens.
 *
 * Summary: Estimate the message tokens appropriately based on current system conditions.
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
export function estimateMessageTokens(messages: any[]): number {
    return messages.reduce((acc, msg) => {
        let content = typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content || "");
        if (msg.toolName) content += ` ${msg.toolName}`;
        if (msg.toolArgs) content += ` ${JSON.stringify(msg.toolArgs)}`;
        if (msg.toolResult) content += ` ${JSON.stringify(msg.toolResult)}`;
        return acc + estimateTokens(content);
    }, 0);
}

/**
 * formatTokenCount serves as a public interface for interacting with formatTokenCount.
 *
 * Summary: Format the token count appropriately based on current system conditions.
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
export function formatTokenCount(count: number): string {
    if (count >= 1000) {
        return (count / 1000).toFixed(1) + 'k';
    }
    return count.toString();
}

/**
 * calculateCost serves as a public interface for interacting with calculateCost.
 *
 * Summary: Calculate the cost appropriately based on current system conditions.
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
export function calculateCost(tokens: number): number {
    // Generic blended rate: $5 per 1M tokens ($0.005 per 1k)
    // This is roughly average for GPT-4o input/output blend or Claude 3.5 Sonnet.
    const RATE_PER_1K = 0.005;
    return (tokens / 1000) * RATE_PER_1K;
}

/**
 * formatCost serves as a public interface for interacting with formatCost.
 *
 * Summary: Format the cost appropriately based on current system conditions.
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
export function formatCost(cost: number): string {
    if (cost === 0) return "$0.00";
    if (cost < 0.01) return `$${cost.toFixed(4)}`;
    return `$${cost.toFixed(2)}`;
}
