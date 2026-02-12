
import { describe, it, expect } from 'vitest';
import { mapRegistryToTemplate, mapCommunityToTemplate, resolveIcon } from './catalog-mapper';
import { ServiceRegistryItem } from './service-registry';
import { CommunityServer } from './marketplace-service';
import { Database, Box } from 'lucide-react';

describe('catalog-mapper', () => {
    it('should map registry item to template correctly', () => {
        const item: ServiceRegistryItem = {
            id: 'test-db',
            name: 'Test DB',
            repo: 'test/repo',
            command: 'npx test-db',
            description: 'A test database',
            configurationSchema: {
                properties: {
                    DB_URL: {
                        title: 'Database URL',
                        description: 'The connection string',
                        default: 'postgres://localhost',
                        format: 'password'
                    }
                }
            }
        };

        const template = mapRegistryToTemplate(item);

        expect(template.id).toBe('registry-test-db');
        expect(template.name).toBe('Test DB');
        // @ts-ignore
        expect(template.source).toBe('verified');
        expect(template.fields).toHaveLength(1);
        expect(template.fields![0]).toEqual({
            name: 'DB_URL',
            label: 'Database URL',
            placeholder: 'The connection string',
            key: 'commandLineService.env.DB_URL',
            defaultValue: 'postgres://localhost',
            type: 'password'
        });
    });

    it('should map community server to template correctly', () => {
        const server: CommunityServer = {
            name: 'My Cool Tool',
            description: 'Does cool things',
            url: 'https://github.com/user/my-cool-tool',
            tags: ['python', 'ai'],
            category: 'Coolness'
        };

        const template = mapCommunityToTemplate(server);

        expect(template.id).toBe('community-my-cool-tool');
        expect(template.name).toBe('My Cool Tool');
        // @ts-ignore
        expect(template.source).toBe('community');
        expect(template.config.commandLineService!.command).toContain('uvx my-cool-tool');
    });

    it('should resolve icons correctly', () => {
        const icon1 = resolveIcon('postgres', 'registry');
        expect(icon1).toBe(Database); // Based on implementation

        const icon2 = resolveIcon('unknown', 'community');
        expect(icon2).toBe(Box);
    });
});
