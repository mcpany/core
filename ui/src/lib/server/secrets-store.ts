/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import crypto from 'crypto';

// Simple in-memory mock store for secrets
// In production this would interact with the backend service via gRPC or DB

// WARNING: This is a fixed key for DEMO purposes only.
// In production, use a key management system (KMS) or environment variables.
// This ensures that even in this mock store, we demonstrate using strong encryption algorithms.
const MOCK_ENCRYPTION_KEY = crypto.scryptSync('mcp-any-demo-secret', 'salt', 32);
const ALGORITHM = 'aes-256-cbc';
const IV_LENGTH = 16;

function encrypt(text: string): string {
    const iv = crypto.randomBytes(IV_LENGTH);
    const cipher = crypto.createCipheriv(ALGORITHM, MOCK_ENCRYPTION_KEY, iv);
    let encrypted = cipher.update(text);
    encrypted = Buffer.concat([encrypted, cipher.final()]);
    return iv.toString('hex') + ':' + encrypted.toString('hex');
}

// Note: Decryption helper for future use or verification
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function decrypt(text: string): string {
    try {
        const textParts = text.split(':');
        if (textParts.length < 2) return "INVALID_FORMAT";
        const iv = Buffer.from(textParts.shift()!, 'hex');
        const encryptedText = Buffer.from(textParts.join(':'), 'hex');
        const decipher = crypto.createDecipheriv(ALGORITHM, MOCK_ENCRYPTION_KEY, iv);
        let decrypted = decipher.update(encryptedText);
        decrypted = Buffer.concat([decrypted, decipher.final()]);
        return decrypted.toString();
    } catch (e) {
        console.error("Failed to decrypt secret", e);
        return "DECRYPTION_FAILED";
    }
}

interface Secret {
    id: string;
    name: string;
    key: string;
    encryptedValue: string;
    provider: string;
    createdAt: string;
    lastUsed: string;
}

let mockSecrets: Secret[] = [
    // Pre-encrypted placeholders (conceptually)
    { id: "sec-1", name: "OpenAI API Key", key: "OPENAI_API_KEY", encryptedValue: "e2b1...", provider: "openai", createdAt: "2024-01-01T00:00:00Z", lastUsed: "2024-05-10T12:00:00Z" },
    { id: "sec-2", name: "AWS Access Key", key: "AWS_ACCESS_KEY_ID", encryptedValue: "a4f2...", provider: "aws", createdAt: "2024-02-15T00:00:00Z", lastUsed: "2024-05-14T09:30:00Z" }
];

export const SecretsStore = {
    getAllDecrypted: () => {
        // In reality, we wouldn't return decrypted values in list
        return {
            secrets: mockSecrets.map(s => ({
                ...s,
                value: "********" // Masked
            }))
        };
    },
    add: (secret: any) => {
        const newSecret = {
            ...secret,
            // SECURITY: Use AES-256 encryption instead of base64 encoding
            encryptedValue: encrypt(secret.value),
            value: undefined
        };
        mockSecrets.push(newSecret);
        return newSecret;
    },
    delete: (id: string) => {
        mockSecrets = mockSecrets.filter(s => s.id !== id);
    }
};
