/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Simple in-memory mock store for secrets
// In production this would interact with the backend service via gRPC or DB

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
    { id: "sec-1", name: "OpenAI API Key", key: "OPENAI_API_KEY", encryptedValue: "sk-...", provider: "openai", createdAt: "2024-01-01T00:00:00Z", lastUsed: "2024-05-10T12:00:00Z" },
    { id: "sec-2", name: "AWS Access Key", key: "AWS_ACCESS_KEY_ID", encryptedValue: "AKIA...", provider: "aws", createdAt: "2024-02-15T00:00:00Z", lastUsed: "2024-05-14T09:30:00Z" }
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
            encryptedValue: Buffer.from(secret.value).toString('base64'),
            value: undefined
        };
        mockSecrets.push(newSecret);
        return newSecret;
    },
    delete: (id: string) => {
        mockSecrets = mockSecrets.filter(s => s.id !== id);
    }
};
