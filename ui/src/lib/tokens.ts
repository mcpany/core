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

    // Simple heuristic used by many LLM providers for estimation:
    // Approximately 4 characters per token for English text.
    // We add some overhead for whitespace and special characters.
    const charCount = text.length;

    // ⚡ BOLT: Optimization - Count words in single pass without allocation.
    // Randomized Selection from Top 5 High-Impact Targets
    // Replaces: const wordCount = text.trim().split(/\s+/).length;
    let wordCount = 0;
    let inWord = false;

    for (let i = 0; i < charCount; i++) {
        const code = text.charCodeAt(i);
        // Common whitespace: Space(32), Tab(9), LF(10), CR(13), NBSP(160)
        // This covers the vast majority of cases.
        // We trade full unicode compliance for raw speed and zero allocation.
        const isSpace = (code === 32 || code === 9 || code === 10 || code === 13 || code === 160);

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
// eslint-disable-next-line @typescript-eslint/no-explicit-any
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
