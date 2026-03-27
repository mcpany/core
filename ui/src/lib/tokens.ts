/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Intent: Document estimateTokens
 *
 * Params:
 *   - None
 *
 * Estimates the number of tokens in a string or object using a simple heuristic.
 *
 * Summary: Provides a rough estimation of the number of tokens for UI purposes.
 *
 * Parameters:
 *   - input (any): The text or object to estimate tokens for.
 *
 * Returns:
 *   - number: The estimated token count based on heuristics.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
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
 * Intent: Document estimateMessageTokens
 *
 * Params:
 *   - None
 *
 * Calculates total tokens for a sequence of messages.
 *
 * Summary: Aggregates the estimated token count for an array of message objects.
 *
 * Parameters:
 *   - messages (any[]): Array of message objects containing content.
 *
 * Returns:
 *   - number: Total estimated tokens for all messages.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
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
 * Intent: Document formatTokenCount
 *
 * Params:
 *   - None
 *
 * Formats a number of tokens into a human-readable string.
 *
 * Summary: Converts a token count into a formatted string (e.g., 1.2k).
 *
 * Parameters:
 *   - count (number): The number of tokens.
 *
 * Returns:
 *   - string: The formatted token count string.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export function formatTokenCount(count: number): string {
    if (count >= 1000) {
        return (count / 1000).toFixed(1) + 'k';
    }
    return count.toString();
}

/**
 * Intent: Document calculateCost
 *
 * Params:
 *   - None
 *
 * Calculates the estimated cost for a given number of tokens.
 *
 * Summary: Estimates the cost in USD based on a generic pricing model.
 *
 * Parameters:
 *   - tokens (number): The number of tokens.
 *
 * Returns:
 *   - number: The estimated cost in USD.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export function calculateCost(tokens: number): number {
    // Generic blended rate: $5 per 1M tokens ($0.005 per 1k)
    // This is roughly average for GPT-4o input/output blend or Claude 3.5 Sonnet.
    const RATE_PER_1K = 0.005;
    return (tokens / 1000) * RATE_PER_1K;
}

/**
 * Intent: Document formatCost
 *
 * Params:
 *   - None
 *
 * Formats a cost into a currency string.
 *
 * Summary: Formats a numerical cost into a USD currency string.
 *
 * Parameters:
 *   - cost (number): The cost in USD.
 *
 * Returns:
 *   - string: The formatted string (e.g., "$0.0024").
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export function formatCost(cost: number): string {
    if (cost === 0) return "$0.00";
    if (cost < 0.01) return `$${cost.toFixed(4)}`;
    return `$${cost.toFixed(2)}`;
}
