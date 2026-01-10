/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { GET, POST } from '@/app/api/secrets/route';
import { DELETE } from '@/app/api/secrets/[id]/route';
import { SecretsStore } from '@/lib/server/secrets-store';

// Mock the store
vi.mock('@/lib/server/secrets-store', () => ({
    SecretsStore: {
        getAllDecrypted: vi.fn(),
        add: vi.fn(),
        delete: vi.fn(),
    }
}));

describe('Secrets API', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    describe('GET /api/secrets', () => {
        it('should return list of secrets', async () => {
            const mockSecrets = [
                { id: '1', name: 'Test', key: 'TEST_KEY', value: 'secret', provider: 'custom' }
            ];
            vi.mocked(SecretsStore.getAllDecrypted).mockReturnValue(mockSecrets as any);

            const response = await GET();
            const data = await response.json();

            expect(response.status).toBe(200);
            expect(data).toEqual(mockSecrets);
            expect(SecretsStore.getAllDecrypted).toHaveBeenCalled();
        });

        it('should handle errors', async () => {
            vi.mocked(SecretsStore.getAllDecrypted).mockImplementation(() => {
                throw new Error('Database error');
            });

            const response = await GET();
            expect(response.status).toBe(500);
        });
    });

    describe('POST /api/secrets', () => {
        it('should create a new secret', async () => {
            const newSecret = { name: 'New', key: 'NEW_KEY', value: 'value123', provider: 'custom' };
            vi.mocked(SecretsStore.add).mockReturnValue({ ...newSecret, id: '123', createdAt: '', lastUsed: '', encryptedValue: 'enc' } as any);

            const req = new Request('http://localhost/api/secrets', {
                method: 'POST',
                body: JSON.stringify(newSecret)
            });

            const response = await POST(req);
            const data = await response.json();

            expect(response.status).toBe(200);
            expect(data.id).toBe('123');
            // Check if value is returned decrypted as per our implementation
            expect(data.value).toBe('value123');
            expect(SecretsStore.add).toHaveBeenCalledWith(expect.objectContaining({
                name: 'New',
                key: 'NEW_KEY',
                value: 'value123'
            }));
        });

        it('should validate input', async () => {
            const req = new Request('http://localhost/api/secrets', {
                method: 'POST',
                body: JSON.stringify({ name: 'Test' }) // Missing key/value
            });

            const response = await POST(req);
            expect(response.status).toBe(400);
        });
    });

    describe('DELETE /api/secrets/[id]', () => {
        it('should delete a secret', async () => {
            const req = new Request('http://localhost/api/secrets/123', {
                method: 'DELETE'
            });
            const params = Promise.resolve({ id: '123' });

            const response = await DELETE(req, { params });
            const data = await response.json();

            expect(response.status).toBe(200);
            expect(data.success).toBe(true);
            expect(SecretsStore.delete).toHaveBeenCalledWith('123');
        });
    });
});
