/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface User {
  id: string;
  roles: string[];
  profile_ids?: string[];
  authentication?: {
    basic_auth?: {
        password_hash?: string;
    };
    api_key?: {
        param_name?: string;
        in?: number;
        verification_value?: string;
    };
  };
}
