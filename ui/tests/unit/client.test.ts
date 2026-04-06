import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from '../../src/lib/client';

global.fetch = vi.fn();

describe('apiClient', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    describe('executePrompt', () => {
        it('calls the correct endpoint with encodeURIComponent for prompt name', async () => {
            const mockResponse = { ok: true, json: async () => ({ result: 'success' }) };
            (global.fetch as any).mockResolvedValue(mockResponse);

            const promptName = 'test prompt with spaces/and/slashes';
            const args = { city: 'London' };

            await apiClient.executePrompt(promptName, args);

            expect(global.fetch).toHaveBeenCalledWith(
                `/api/v1/prompts/test%20prompt%20with%20spaces%2Fand%2Fslashes/execute`,
                expect.objectContaining({
                    method: 'POST',
                    body: JSON.stringify(args)
                })
            );
        });

        it('throws an error if the response is not ok', async () => {
            const mockResponse = { ok: false, status: 500 };
            (global.fetch as any).mockResolvedValue(mockResponse);

            await expect(apiClient.executePrompt('test', {})).rejects.toThrow('Failed to execute prompt');
        });
    });
});
