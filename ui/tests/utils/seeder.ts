
import { apiClient, UpstreamServiceConfig } from '@/lib/client';

export class Seeder {
  /**
   * Registers a service, ensuring any existing service with the same name is removed first.
   */
  static async registerService(config: UpstreamServiceConfig) {
    const serviceName = config.name || config.id;
    if (!serviceName) {
        throw new Error('Service config must have a name or id');
    }

    // Always attempt to delete first to ensure clean state
    // We ignore errors here because the service might not exist
    try {
        console.log(`[Seeder] Ensuring ${serviceName} is clean...`);
        await apiClient.unregisterService(serviceName);
        // Small delay to allow backend to propagate deletion if async
        await new Promise(r => setTimeout(r, 500));
    } catch (e) {
        // Ignore delete errors
    }

    console.log(`[Seeder] Registering service ${serviceName}...`);
    return await apiClient.registerService(config);
  }

  /**
   * Cleans up all registered services.
   */
  static async cleanup() {
      console.log('[Seeder] Cleaning up all services...');
      try {
          const services = await apiClient.listServices();
          for (const s of services) {
              const name = s.name || s.id;
              if (name) {
                  console.log(`[Seeder] Deleting service ${name}...`);
                  await apiClient.unregisterService(name);
              }
          }
      } catch (e) {
          console.error("[Seeder] Cleanup failed:", e);
      }
  }
}
