    /**
     * Lists all profiles.
     */
    listProfiles: async () => {
        const res = await fetchWithAuth('/api/v1/profiles');
        if (!res.ok) throw new Error('Failed to list profiles');
        const data = await res.json();
        return data.profiles || [];
    },
