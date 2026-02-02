import { apiClient } from './client';

describe('client', () => {
    describe('getCurrentUserId', () => {
        beforeEach(() => {
            localStorage.clear();
        });

        it('returns null if no token', () => {
            expect(apiClient.getCurrentUserId()).toBeNull();
        });

        it('returns username from valid token', () => {
            const token = btoa('user1:pass1');
            localStorage.setItem('mcp_auth_token', token);
            expect(apiClient.getCurrentUserId()).toBe('user1');
        });

        it('handles malformed token gracefully', () => {
            localStorage.setItem('mcp_auth_token', 'not-base64');
            // atob might throw, implementation catches it
            const result = apiClient.getCurrentUserId();
            expect(result).toBeNull();
        });
    });
});
