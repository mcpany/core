    /**
     * Lists all available service templates.
     */
    listTemplates: async () => {
        return apiClient.getServiceTemplates();
    },

    /**
     * Saves a service template.
     * @param config The service configuration to save as a template.
     */
    saveTemplate: async (config: UpstreamServiceConfig) => {
        const payload = {
            name: config.name,
            description: config.description || "Custom Template",
            service_config: {
                ...config,
                // Map back to snake_case if needed by backend, but let's assume
                // registerService mapping logic applies.
                // Actually, the templates endpoint usually expects a specific format.
                // Let's reuse the mapping from registerService but wrap it.
                // For simplicity, we send the config object and let the server handle it or fail if it expects strict snake_case.
                // Given the existing code uses fetchWithAuth, we'll try to match that pattern.
                id: config.id,
                name: config.name,
            }
        };
        // We need to map camelCase to snake_case for the backend
        const mappedConfig: any = {
            id: config.id,
            name: config.name,
            version: config.version,
            command_line_service: config.commandLineService ? {
                command: config.commandLineService.command,
                working_directory: config.commandLineService.workingDirectory,
                env: config.commandLineService.env
            } : undefined,
            openapi_service: config.openapiService ? {
                spec_url: config.openapiService.specUrl,
                spec_content: config.openapiService.specContent
            } : undefined,
            // ... other fields as needed
        };

        const res = await fetchWithAuth('/api/v1/templates', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                name: config.name,
                description: config.description,
                service_config: mappedConfig
            })
        });
        if (!res.ok) throw new Error('Failed to save template');
        return res.json();
    },

    /**
     * Deletes a service template.
     * @param id The ID of the template.
     */
    deleteTemplate: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/templates/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete template');
        return {};
    },
