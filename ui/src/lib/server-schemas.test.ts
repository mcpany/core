/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { getSchemaForServer } from './server-schemas';

describe('getSchemaForServer', () => {
  it('should return correct schema for exact match (GitHub)', () => {
    const schema = getSchemaForServer('modelcontextprotocol', 'server-github');
    expect(schema).not.toBeNull();
    const parsed = JSON.parse(schema!);
    expect(parsed.properties).toHaveProperty('GITHUB_PERSONAL_ACCESS_TOKEN');
  });

  it('should return correct schema for exact match (Cloudflare)', () => {
    const schema = getSchemaForServer('modelcontextprotocol', 'server-cloudflare');
    expect(schema).not.toBeNull();
    const parsed = JSON.parse(schema!);
    expect(parsed.properties).toHaveProperty('CLOUDFLARE_API_TOKEN');
  });

  it('should return correct schema for generic match (mcp-server-postgres -> postgres)', () => {
    // Testing the fallback logic
    const schema = getSchemaForServer('someuser', 'mcp-server-postgres');
    expect(schema).not.toBeNull();
    const parsed = JSON.parse(schema!);
    expect(parsed.properties).toHaveProperty('POSTGRES_URL');
  });

  it('should return correct schema for generic match (server-slack -> slack)', () => {
    const schema = getSchemaForServer('someuser', 'server-slack');
    expect(schema).not.toBeNull();
    const parsed = JSON.parse(schema!);
    expect(parsed.properties).toHaveProperty('SLACK_BOT_TOKEN');
  });

  it('should return null for unknown server', () => {
    const schema = getSchemaForServer('someuser', 'unknown-server');
    expect(schema).toBeNull();
  });

  it('should handle case insensitivity', () => {
    const schema = getSchemaForServer('ModelContextProtocol', 'SERVER-GITHUB');
    expect(schema).not.toBeNull();
  });
});
