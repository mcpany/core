
// Manually defined types matching the proto definitions for UI use

export interface UpstreamServiceConfig {
  id: string;
  name: string;
  version?: string;
  disable: boolean;
  priority: number;
  connectionPool?: ConnectionPoolConfig;
  serviceConfig?: {
    case: string;
    value: any;
  };
}

export interface ConnectionPoolConfig {
  maxConnections: number;
  maxIdleConnections: number;
  idleTimeout: {
    seconds: number;
  };
}
