/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Gets the current user ID from the stored auth token.
 * Assumes the token is a Basic Auth base64 encoded string (username:password).
 * @returns The username (which serves as ID) or null if not authenticated.
 */
export function getCurrentUserId(): string | null {
    if (typeof window === "undefined") return null;

    const token = localStorage.getItem("mcp_auth_token");
    if (!token) return null;

    try {
        const decoded = atob(token);
        const parts = decoded.split(":");
        if (parts.length >= 1) {
            return parts[0];
        }
    } catch (e) {
        console.error("Failed to decode auth token", e);
    }
    return null;
}
