/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Estimates the number of tokens in a string using a simple heuristic.
 *
 * Summary: Approximates token count for a given text.
 *
 * Description:
 * This function uses a heuristic based on character and word counts to estimate
 * the number of tokens. It is intended for UI feedback and is not a precise
 * tokenizer (like tiktoken). It assumes approximately 4 characters or 1.3 words
 * per token.
 *
 * @param text - The text to estimate tokens for.
 * @returns The estimated number of tokens.
 */
export function estimateTokens(text: string): number {
    if (!text) return 0;

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
 * Calculates total tokens for a sequence of messages.
 *
 * Summary: Sums the estimated token counts for a list of messages.
 *
 * Description:
 * Iterates through a list of message objects, stringifying content and tool
 * arguments as needed, and aggregates the estimated token count.
 *
 * @param messages - Array of message objects (e.g., from an LLM conversation).
 * @returns The total estimated token count.
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
 * Formats a number of tokens into a human-readable string.
 *
 * Summary: Formats a token count with 'k' suffix if applicable.
 *
 * @param count - The raw number of tokens.
 * @returns A formatted string (e.g., "1.2k" or "500").
 */
export function formatTokenCount(count: number): string {
    if (count >= 1000) {
        return (count / 1000).toFixed(1) + 'k';
    }
    return count.toString();
}
