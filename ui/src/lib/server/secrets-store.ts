/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import path from 'path';
import crypto from 'crypto';

// Define the shape of our secret
export interface Secret {
    id: string;
    name: string;
    key: string;
    encryptedValue: string; // Stored encrypted
    provider: string;
    createdAt: string;
    lastUsed: string;
}

const STORAGE_FILE = path.join(process.cwd(), 'secrets-store.json');
const MASTER_KEY_FILE = path.join(process.cwd(), 'secrets-master.key');

// Get or create the master key for encryption
function getMasterKey(): Buffer {
    if (process.env.MCPANY_MASTER_KEY) {
        // Use provided env key (must be 32 bytes hex or suitable length, here we hash it to ensure 32 bytes)
        return crypto.createHash('sha256').update(process.env.MCPANY_MASTER_KEY).digest();
    }

    if (fs.existsSync(MASTER_KEY_FILE)) {
        try {
            const keyHex = fs.readFileSync(MASTER_KEY_FILE, 'utf-8').trim();
            return Buffer.from(keyHex, 'hex');
        } catch (_e) {
            // Fallback to generating a new one if file is corrupt
            console.error('Failed to read master key file, generating new one. WARNING: Existing secrets will be lost.');
        }
    }

    // Generate new key
    const newKey = crypto.randomBytes(32);
    try {
        fs.writeFileSync(MASTER_KEY_FILE, newKey.toString('hex'), { mode: 0o600 });
    } catch (_e) {
        console.error('Failed to write master key file:', _e);
    }
    return newKey;
}

const MASTER_KEY = getMasterKey();
const ALGORITHM = 'aes-256-gcm';

// Helper to ensure file exists
function ensureStore() {
    if (!fs.existsSync(STORAGE_FILE)) {
        fs.writeFileSync(STORAGE_FILE, JSON.stringify([]));
    }
}

// Secure encryption using AES-256-GCM
function encrypt(text: string): string {
    const iv = crypto.randomBytes(12); // 96-bit IV for GCM
    const cipher = crypto.createCipheriv(ALGORITHM, MASTER_KEY, iv);

    let encrypted = cipher.update(text, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    const authTag = cipher.getAuthTag().toString('hex');

    // Format: v1:iv:authTag:encrypted
    return `v1:${iv.toString('hex')}:${authTag}:${encrypted}`;
}

function decrypt(text: string): string {
    // Check for version prefix
    if (text.startsWith('v1:')) {
        try {
            const parts = text.split(':');
            if (parts.length !== 4) throw new Error('Invalid format');

            const iv = Buffer.from(parts[1], 'hex');
            const authTag = Buffer.from(parts[2], 'hex');
            const encrypted = parts[3];

            const decipher = crypto.createDecipheriv(ALGORITHM, MASTER_KEY, iv);
            decipher.setAuthTag(authTag);

            let decrypted = decipher.update(encrypted, 'hex', 'utf8');
            decrypted += decipher.final('utf8');
            return decrypted;
        } catch (_e) {
            console.error('Decryption failed:', _e);
            return '[Decryption Failed]';
        }
    }

    // Fallback: Try Legacy Base64
    try {
        const decoded = Buffer.from(text, 'base64').toString('utf-8');
        // Simple heuristic: if it looks like garbage or fails, we might return error,
        // but since we are migrating from Base64, we assume valid Base64 string if not v1.
        return decoded;
    } catch (_e) {
        return '[Invalid Secret]';
    }
}

export const SecretsStore = {
    getAll: (): Secret[] => {
        ensureStore();
        try {
            const data = fs.readFileSync(STORAGE_FILE, 'utf-8');
            return JSON.parse(data);
        } catch (_e) {
            return [];
        }
    },

    saveAll: (secrets: Secret[]) => {
        fs.writeFileSync(STORAGE_FILE, JSON.stringify(secrets, null, 2));
    },

    add: (secret: Omit<Secret, 'encryptedValue'> & { value: string }) => {
        ensureStore();
        const secrets = SecretsStore.getAll();
        const newSecret: Secret = {
            id: secret.id,
            name: secret.name,
            key: secret.key,
            provider: secret.provider,
            createdAt: secret.createdAt,
            lastUsed: secret.lastUsed,
            encryptedValue: encrypt(secret.value)
        };
        secrets.push(newSecret);
        SecretsStore.saveAll(secrets);
        return newSecret;
    },

    delete: (id: string) => {
        ensureStore();
        let secrets = SecretsStore.getAll();
        secrets = secrets.filter(s => s.id !== id);
        SecretsStore.saveAll(secrets);
    },

    // Used for returning to UI (decrypted)
    // Also handles migration from legacy format
    getAllDecrypted: () => {
        const secrets = SecretsStore.getAll();
        let migrationNeeded = false;

        const decryptedSecrets = secrets.map(s => {
            let value = decrypt(s.encryptedValue);

            // Check if this was a legacy secret (didn't start with v1:)
            if (!s.encryptedValue.startsWith('v1:') && value !== '[Decryption Failed]') {
                // It was successfully decrypted as legacy, verify if we should migrate
                // We re-encrypt immediately
                s.encryptedValue = encrypt(value);
                migrationNeeded = true;
            }

            return {
                ...s,
                value: value,
                encryptedValue: undefined // Don't send this back
            };
        });

        if (migrationNeeded) {
            // Save the migrated secrets back to disk (with new encryption)
            // Note: decryptedSecrets has `value` and `undefined` encryptedValue, so we can't save it directly.
            // We need to save the `secrets` array which we modified in place above.
            SecretsStore.saveAll(secrets);

        }

        return decryptedSecrets;
    }
};
