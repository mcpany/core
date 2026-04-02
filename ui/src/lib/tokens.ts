/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Summary: Provides a rough estimation of the number of tokens for UI purposes using standard English heuristics.
 *
 * Params:
 *   - input (any): The text or object to estimate tokens for. If an object is provided, it will be JSON stringified.
 *
 * Returns:
 *   - number: The estimated token count based on character and word count heuristics.
 *
 * Errors:
 *   - Throws standard JSON.stringify errors if the object contains circular references.
 *
 * Side Effects:
 *   - Performs no state mutations. Calculates locally.
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
 * Summary: Aggregates the estimated token count for an array of multimodal message objects.
 *
 * Params:
 *   - messages (any[]): Array of message objects containing textual content, tool names, arguments, and results.
 *
 * Returns:
 *   - number: Total estimated tokens for all messages, including tool context.
 *
 * Errors:
 *   - Throws standard JSON.stringify errors if message objects contain circular references.
 *
 * Side Effects:
 *   - Performs no state mutations. Evaluates array synchronously.
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
 * Summary: Converts a raw token count into a human-readable abbreviated string (e.g., 1.2k).
 *
 * Params:
 *   - count (number): The raw integer number of tokens to format.
 *
 * Returns:
 *   - string: The formatted token count string, safely scaled to 'k' for large numbers.
 *
 * Errors:
 *   - N/A: Pure formatting function.
 *
 * Side Effects:
 *   - Performs no state mutations.
 */
export function formatTokenCount(count: number): string {
    if (count >= 1000) {
        return (count / 1000).toFixed(1) + 'k';
    }
    return count.toString();
}

/**
 * Summary: Estimates the cost in USD based on a generic blended rate pricing model.
 *
 * Params:
 *   - tokens (number): The integer number of tokens to calculate the cost for.
 *
 * Returns:
 *   - number: The estimated cost in floating-point USD.
 *
 * Errors:
 *   - N/A: Pure mathematical function.
 *
 * Side Effects:
 *   - Performs no state mutations.
 */
export function calculateCost(tokens: number): number {
    // Generic blended rate: $5 per 1M tokens ($0.005 per 1k)
    // This is roughly average for GPT-4o input/output blend or Claude 3.5 Sonnet.
    const RATE_PER_1K = 0.005;
    return (tokens / 1000) * RATE_PER_1K;
}

/**
 * Summary: Formats a numerical cost floating-point into a proper USD currency string.
 *
 * Params:
 *   - cost (number): The raw cost float in USD.
 *
 * Returns:
 *   - string: The safely formatted currency string (e.g., "$0.0024" or "$0.00").
 *
 * Errors:
 *   - N/A: Pure formatting function.
 *
 * Side Effects:
 *   - Performs no state mutations.
 */
export function formatCost(cost: number): string {
    if (cost === 0) return "$0.00";
    if (cost < 0.01) return `$${cost.toFixed(4)}`;
    return `$${cost.toFixed(2)}`;
}
