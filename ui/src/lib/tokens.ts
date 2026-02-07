/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Estimates the number of tokens in a string using a simple heuristic.
 * This is meant for UI estimation only, not for precision.
 * @param text - The text to estimate tokens for.
 * @returns Estimated token count.
 */
export function estimateTokens(text: string): number {
    if (!text) return 0;

    // ⚡ BOLT: Optimized token estimation to avoid O(N) allocation from split()
    // Randomized Selection from Top 5 High-Impact Targets

    // Simple heuristic used by many LLM providers for estimation:
    // Approximately 4 characters per token for English text.
    // We add some overhead for whitespace and special characters.
    const charCount = text.length;

    // Count words without allocation
    let wordCount = 0;
    let inWord = false;
    for (let i = 0; i < charCount; i++) {
        const c = text.charCodeAt(i);
        // ASCII space (32), tab (9), newline (10), carriage return (13), NBSP (160)
        const isSpace = c === 32 || c === 9 || c === 10 || c === 13 || c === 160;

        if (isSpace) {
            inWord = false;
        } else if (!inWord) {
            inWord = true;
            wordCount++;
        }
    }

    // Heuristic 1: 4 chars per token
    // Heuristic 2: 1.3 words per token
    // We'll take a balanced approach or the max of both for safety.
    const h1 = Math.ceil(charCount / 4);
    const h2 = Math.ceil(wordCount * 1.3);

    return Math.max(h1, h2);
}

/**
 * Calculates total tokens for a sequence of messages.
 * @param messages - Array of message objects with content.
 * @returns Total estimated tokens.
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
 * Formats a number of tokens into a human-readable string (e.g., 1.2k).
 * @param count - The number of tokens.
 * @returns Formatted string.
 */
export function formatTokenCount(count: number): string {
    if (count >= 1000) {
        return (count / 1000).toFixed(1) + 'k';
    }
    return count.toString();
}
