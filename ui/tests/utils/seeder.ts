import { apiClient, UpstreamServiceConfig } from '@/lib/client';

export class Seeder {
    private createdServices: string[] = [];

    /**
     * Registers a service using the API client.
     * @param config The service configuration to register.
     */
    /**
     * Manually track a service for cleanup.
     * Use this when the service is created by the UI or other means.
     * Should be the Service ID if possible, otherwise Name (if ID=Name).
     */
    trackService(idOrName: string) {
        if (!this.createdServices.includes(idOrName)) {
            this.createdServices.push(idOrName);
        }
    }

    async registerService(config: UpstreamServiceConfig) {
        console.log(`Seeding service: ${config.name} (${config.id})`);

        // Pre-cleanup: Try to unregister if it exists to ensure clean state
        // This handles cases where a previous test run crashed without cleanup
        const id = config.id || config.name;
        if (id) {
            try {
                await apiClient.unregisterService(id);
                console.log(`Pre-cleaned service: ${id}`);
            } catch (e) {
                // Ignore 404s or other errors during pre-cleanup
            }
        }

        try {
            // Ensure ID is set for deterministic cleanup
            if (!config.id) {
                config.id = config.name; // Fallback
            }
            await apiClient.registerService(config);
            // Track ID for cleanup, as unregisterService expects ID
            const idToTrack = config.id || config.name;
            if (idToTrack) {
                this.createdServices.push(idToTrack);
            }
        } catch (e: any) {
            if (e.message && e.message.includes("already exists")) {
                console.log(`Service ${config.name} already exists, attempting update...`);
                try {
                    await apiClient.updateService(config);
                    const idToTrack = config.id || config.name;
                    if (idToTrack && !this.createdServices.includes(idToTrack)) {
                        this.createdServices.push(idToTrack);
                    }
                } catch (updateErr) {
                    console.error(`Failed to update service ${config.name}:`, updateErr);
                    throw updateErr;
                }
            } else {
                console.error(`Failed to register service ${config.name}:`, e);
                throw e;
            }
        }
    }

    /**
     * Cleans up all services created by this seeder.
     */
    async cleanup() {
        console.log(`Cleaning up ${this.createdServices.length} seeded services...`);
        // Reverse order cleanup
        for (let i = this.createdServices.length - 1; i >= 0; i--) {
            const name = this.createdServices[i];
            try {
                // We use name as ID often, but let's check client implementation.
                // client.ts unregisterService takes 'id'.
                // If we used name as ID in registerService, we should use it here.
                await apiClient.unregisterService(name);
                console.log(`Unregistered service: ${name}`);
            } catch (e) {
                console.warn(`Failed to unregister service ${name}:`, e);
            }
        }
        this.createdServices = [];
    }
}
