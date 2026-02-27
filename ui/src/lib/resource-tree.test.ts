/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { buildResourceTree, TreeNode } from './resource-tree';
import { ResourceDefinition } from './client';
import { describe, it, expect } from 'vitest';

describe('buildResourceTree', () => {
  it('should handle file:// URIs correctly', () => {
    const resources: ResourceDefinition[] = [
      { uri: 'file:///app/src/main.go', name: 'main.go', mimeType: 'text/x-go' },
      { uri: 'file:///app/public/index.html', name: 'index.html', mimeType: 'text/html' }
    ];

    const tree = buildResourceTree(resources);

    // Expected structure:
    // - app (folder)
    //   - src (folder)
    //     - main.go (file)
    //   - public (folder)
    //     - index.html (file)

    expect(tree).toHaveLength(1);
    const app = tree[0];
    expect(app.name).toBe('app');
    expect(app.type).toBe('folder');
    expect(app.children).toHaveLength(2);

    const publicFolder = app.children?.find(c => c.name === 'public');
    expect(publicFolder).toBeDefined();
    expect(publicFolder?.children).toHaveLength(1);
    expect(publicFolder?.children?.[0].name).toBe('index.html');
    expect(publicFolder?.children?.[0].type).toBe('file');

    const srcFolder = app.children?.find(c => c.name === 'src');
    expect(srcFolder).toBeDefined();
    expect(srcFolder?.children).toHaveLength(1);
    expect(srcFolder?.children?.[0].name).toBe('main.go');
  });

  it('should handle other schemes correctly', () => {
    const resources: ResourceDefinition[] = [
        { uri: 'postgres://db/users/schema', name: 'users_schema', mimeType: 'sql' }
    ];

    const tree = buildResourceTree(resources);

    // Expected structure:
    // - postgres:// (folder)
    //   - db (folder)
    //     - users (folder)
    //       - schema (file)

    expect(tree).toHaveLength(1);
    const scheme = tree[0];
    expect(scheme.name).toBe('postgres://');
    expect(scheme.type).toBe('folder');

    expect(scheme.children).toHaveLength(1);
    const db = scheme.children?.[0];
    expect(db?.name).toBe('db');

    const users = db?.children?.[0];
    expect(users?.name).toBe('users');

    const schema = users?.children?.[0];
    expect(schema?.name).toBe('schema');
    expect(schema?.type).toBe('file');
  });

  it('should handle mixed schemes', () => {
     const resources: ResourceDefinition[] = [
        { uri: 'file:///config.json', name: 'config.json', mimeType: 'json' },
        { uri: 's3://bucket/image.png', name: 'image.png', mimeType: 'image/png' }
     ];

     const tree = buildResourceTree(resources);

     // Should have 'config.json' at root (since file:/// has empty parts split behavior check)
     // Wait, file:///config.json -> parts = ["config.json"] -> file node at root.
     // s3://bucket/image.png -> s3:// (folder) -> bucket (folder) -> image.png (file)

     expect(tree).toHaveLength(2);

     const fileNode = tree.find(n => n.name === 'config.json');
     expect(fileNode).toBeDefined();
     expect(fileNode?.type).toBe('file');

     const s3Node = tree.find(n => n.name === 's3://');
     expect(s3Node).toBeDefined();
     expect(s3Node?.type).toBe('folder');
  });
});
