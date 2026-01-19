/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import crypto from 'crypto';

// Simple in-memory mock store for secrets
// In production this would interact with the backend service via gRPC or DB

// 🚨 SECURITY NOTICE:
// This is a MOCK store. We use a random key for the session to ensure
// secrets are encrypted in memory, but they are lost on restart.
// In a real production environment, use a KMS or Vault.
const MOCK_ENCRYPTION_KEY = process.env.MOCK_ENCRYPTION_KEY
    ? crypto.scryptSync(process.env.MOCK_ENCRYPTION_KEY, 'salt', 32)
    : crypto.randomBytes(32);

const ALGORITHM = 'aes-256-gcm';

interface Secret {
    id: string;
    name: string;
    key: string;
    encryptedValue: string;
    iv: string;
    authTag: string;
    provider: string;
    createdAt: string;
    lastUsed: string;
}

// Helper to encrypt data
function encrypt(text: string): { encryptedValue: string, iv: string, authTag: string } {
    const iv = crypto.randomBytes(12);
    const cipher = crypto.createCipheriv(ALGORITHM, MOCK_ENCRYPTION_KEY, iv);
    let encrypted = cipher.update(text, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    return {
        encryptedValue: encrypted,
        iv: iv.toString('hex'),
        authTag: cipher.getAuthTag().toString('hex')
    };
}

// Helper to decrypt data
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function decrypt(encryptedValue: string, iv: string, authTag: string): string {
    const decipher = crypto.createDecipheriv(ALGORITHM, MOCK_ENCRYPTION_KEY, Buffer.from(iv, 'hex'));
    decipher.setAuthTag(Buffer.from(authTag, 'hex'));
    let decrypted = decipher.update(encryptedValue, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    return decrypted;
}

// Initial mock data
const secret1 = encrypt("sk-mock-openai-key-value-must-be-replaced");
const secret2 = encrypt("AKIA-mock-aws-key-value-must-be-replaced");

let mockSecrets: Secret[] = [
    {
        id: "sec-1",
        name: "OpenAI API Key",
        key: "OPENAI_API_KEY",
        encryptedValue: secret1.encryptedValue,
        iv: secret1.iv,
        authTag: secret1.authTag,
        provider: "openai",
        createdAt: "2024-01-01T00:00:00Z",
        lastUsed: "2024-05-10T12:00:00Z"
    },
    {
        id: "sec-2",
        name: "AWS Access Key",
        key: "AWS_ACCESS_KEY_ID",
        encryptedValue: secret2.encryptedValue,
        iv: secret2.iv,
        authTag: secret2.authTag,
        provider: "aws",
        createdAt: "2024-02-15T00:00:00Z",
        lastUsed: "2024-05-14T09:30:00Z"
    }
];

export const SecretsStore = {
    getAllDecrypted: () => {
        // In reality, we wouldn't return decrypted values in list usually,
        // but for this mock we just return the masked version as before.
        return {
            secrets: mockSecrets.map(s => ({
                id: s.id,
                name: s.name,
                key: s.key,
                provider: s.provider,
                createdAt: s.createdAt,
                lastUsed: s.lastUsed,
                value: "********" // Masked
            }))
        };
    },
    add: (secret: any) => {
        const { encryptedValue, iv, authTag } = encrypt(secret.value);
        const newSecret: Secret = {
            id: secret.id || `sec-${Date.now()}`,
            name: secret.name,
            key: secret.key,
            provider: secret.provider,
            createdAt: new Date().toISOString(),
            lastUsed: new Date().toISOString(),
            encryptedValue,
            iv,
            authTag
        };
        mockSecrets.push(newSecret);
        return { ...newSecret, value: "********" };
    },
    delete: (id: string) => {
        mockSecrets = mockSecrets.filter(s => s.id !== id);
    }
};
