/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { SERVICE_TEMPLATES } from './templates';

test('templates are defined', () => {
  expect(SERVICE_TEMPLATES).toBeDefined();
  expect(SERVICE_TEMPLATES.length).toBeGreaterThan(0);
});

test('postgres template has correct config', () => {
  const postgres = SERVICE_TEMPLATES.find(t => t.id === 'postgres');
  expect(postgres).toBeDefined();
  expect(postgres?.config.commandLineService?.command).toContain('server-postgres');
});

test('github template has correct config', () => {
    const github = SERVICE_TEMPLATES.find(t => t.id === 'github');
    expect(github).toBeDefined();
    expect(github?.config.commandLineService?.command).toContain('server-github');
    expect(github?.config.commandLineService?.env?.['GITHUB_PERSONAL_ACCESS_TOKEN']).toBeDefined();
});

test('empty template exists', () => {
    const empty = SERVICE_TEMPLATES.find(t => t.id === 'empty');
    expect(empty).toBeDefined();
    expect(empty?.name).toBe('Custom Service');
});
