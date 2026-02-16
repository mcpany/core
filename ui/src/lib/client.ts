    /**
     * Lists logs from the server (persistent history).
     * @param options Filtering options.
     * @returns A promise that resolves to a list of log entries.
     */
    listLogs: async (options?: { limit?: number; offset?: number; level?: string; source?: string; search?: string }): Promise<any[]> => {
        const query = new URLSearchParams();
        if (options?.limit) query.set('limit', options.limit.toString());
        if (options?.offset) query.set('offset', options.offset.toString());
        if (options?.level && options.level !== 'ALL') query.set('level', options.level);
        if (options?.source && options.source !== 'ALL') query.set('source', options.source);
        if (options?.search) query.set('search', options.search);

        // Retry logic for robustness
        const fetchWithRetry = async (retries = 3, delay = 1000) => {
            for (let i = 0; i < retries; i++) {
                try {
                    const res = await fetchWithAuth(`/api/v1/logs?${query.toString()}`);
                    if (!res.ok) {
                        if (res.status === 404 || res.status === 501) return [];
                        throw new Error('Failed to fetch logs');
                    }
                    return res.json();
                } catch (err) {
                    if (i === retries - 1) throw err;
                    await new Promise(res => setTimeout(res, delay));
                }
            }
        };

        return fetchWithRetry();
    },
