export interface SystemStatus {
  uptime_seconds: number;
  version: string;
  active_connections: number;
  bound_http_port: number;
  bound_grpc_port: number;
  security_warnings: string[];
}
