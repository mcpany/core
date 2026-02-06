/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { analyzeRepository, parseGithubUrl } from './repo-analyzer';

// Mock fetch
const fetchMock = vi.fn();
global.fetch = fetchMock;

describe('repo-analyzer', () => {
    describe('parseGithubUrl', () => {
        it('should parse simple repo url', () => {
            const res = parseGithubUrl('https://github.com/owner/repo');
            expect(res).toEqual({ owner: 'owner', repo: 'repo', branch: undefined, path: '' });
        });

        it('should parse monorepo url with branch', () => {
            const res = parseGithubUrl('https://github.com/owner/repo/tree/main/packages/server');
            expect(res).toEqual({ owner: 'owner', repo: 'repo', branch: 'main', path: 'packages/server' });
        });

        it('should return null for invalid url', () => {
            expect(parseGithubUrl('https://google.com')).toBeNull();
        });
    });

    describe('analyzeRepository', () => {
        beforeEach(() => {
            fetchMock.mockReset();
        });

        it('should handle monorepo path', async () => {
            fetchMock.mockImplementation((url: string) => {
                // Expect URL to contain path
                if (url.includes('packages/server/package.json')) {
                    return Promise.resolve({
                        ok: true,
                        text: () => Promise.resolve(JSON.stringify({ name: '@scope/pkg', description: 'Mono server' })),
                    });
                }
                return Promise.resolve({ ok: false });
            });

            const result = await analyzeRepository('https://github.com/owner/repo/tree/main/packages/server');
            expect(result.detectedType).toBe('node');
            expect(result.name).toBe('@scope/pkg');
            expect(result.command).toBe('npx -y @scope/pkg');
        });

        it('should fallback to master if main fails', async () => {
            fetchMock.mockImplementation((url: string) => {
                if (url.includes('/main/')) return Promise.resolve({ ok: false });
                if (url.includes('/master/package.json')) {
                    return Promise.resolve({
                        ok: true,
                        text: () => Promise.resolve(JSON.stringify({ name: 'legacy-pkg' })),
                    });
                }
                return Promise.resolve({ ok: false });
            });

            const result = await analyzeRepository('https://github.com/owner/repo');
            expect(result.name).toBe('legacy-pkg');
        });
    });
});
