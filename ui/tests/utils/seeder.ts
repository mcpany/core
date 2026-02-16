
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

    try {
        // Try to get the service to see if it exists
        await apiClient.getService(serviceName);
        console.log(`Service ${serviceName} exists, deleting...`);
        await apiClient.unregisterService(serviceName);
    } catch (e) {
        // Service likely doesn't exist or error, proceed
    }

    console.log(`Registering service ${serviceName}...`);
    return await apiClient.registerService(config);
  }

  /**
   * Cleans up all registered services.
   */
  static async cleanup() {
      console.log('Cleaning up all services...');
      try {
          const services = await apiClient.listServices();
          for (const s of services) {
              const name = s.name || s.id;
              if (name) {
                  console.log(`Deleting service ${name}...`);
                  await apiClient.unregisterService(name);
              }
          }
      } catch (e) {
          console.error("Cleanup failed:", e);
          // Don't throw to avoid failing tests during teardown
      }
  }
}
