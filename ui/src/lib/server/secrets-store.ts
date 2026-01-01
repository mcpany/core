/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import path from 'path';

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

// Helper to ensure file exists
function ensureStore() {
    if (!fs.existsSync(STORAGE_FILE)) {
        fs.writeFileSync(STORAGE_FILE, JSON.stringify([]));
    }
}

// Simple encryption (mock) - In real world use standard crypto lib
function encrypt(text: string): string {
    return Buffer.from(text).toString('base64');
}

function decrypt(text: string): string {
    return Buffer.from(text, 'base64').toString('utf-8');
}

export const SecretsStore = {
    getAll: (): Secret[] => {
        ensureStore();
        try {
            const data = fs.readFileSync(STORAGE_FILE, 'utf-8');
            return JSON.parse(data);
        } catch (e) {
            return [];
        }
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
        fs.writeFileSync(STORAGE_FILE, JSON.stringify(secrets, null, 2));
        return newSecret;
    },

    delete: (id: string) => {
        ensureStore();
        let secrets = SecretsStore.getAll();
        secrets = secrets.filter(s => s.id !== id);
        fs.writeFileSync(STORAGE_FILE, JSON.stringify(secrets, null, 2));
    },

    // Used for returning to UI (decrypted)
    getAllDecrypted: () => {
        const secrets = SecretsStore.getAll();
        return secrets.map(s => ({
            ...s,
            value: decrypt(s.encryptedValue),
            encryptedValue: undefined // Don't send this back if we send value
        }));
    }
};
