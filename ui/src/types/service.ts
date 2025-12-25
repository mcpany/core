/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface Service {
  id: string;
  name: string;
  version: string;
  disable: boolean;
  service_config?: {
    http_service?: {
      address: string;
    };
    grpc_service?: {
      address: string;
    };
    [key: string]: any;
  };
  connection_pool?: {
      max_connections: number;
  };
}
