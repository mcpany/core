/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { apiClient } from './client';

// Mock global fetch
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('apiClient', () => {
    beforeEach(() => {
        mockFetch.mockClear();
    });

    describe('getDoctorStatus', () => {
        it('should return doctor report on success', async () => {
            const mockResponse = {
                status: 'ok',
                checks: {
                    configuration: { status: 'ok' }
                }
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            });

            const result = await apiClient.getDoctorStatus();

            expect(mockFetch).toHaveBeenCalledWith('/api/v1/doctor', expect.anything());
            expect(result).toEqual(mockResponse);
        });

        it('should throw error on failure', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 500,
            });

            await expect(apiClient.getDoctorStatus()).rejects.toThrow('Failed to fetch doctor status');
        });
    });
});
