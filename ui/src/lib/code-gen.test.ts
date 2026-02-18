/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { generateCurlCommand, generatePythonCommand } from './code-gen';

describe('code-gen', () => {
  it('generates a valid curl command', () => {
    const toolName = 'test_tool';
    const args = { foo: 'bar', count: 1 };
    const curl = generateCurlCommand(toolName, args);

    expect(curl).toContain('curl -X POST');
    expect(curl).toContain('"name": "test_tool"');
    expect(curl).toContain('"foo": "bar"');
    expect(curl).toContain('"count": 1');
  });

  it('generates a valid python command', () => {
    const toolName = 'test_tool';
    const args = { foo: 'bar', count: 1 };
    const python = generatePythonCommand(toolName, args);

    expect(python).toContain('import requests');
    expect(python).toContain('"name": "test_tool"');
    expect(python).toContain('"arguments": {');
    expect(python).toContain('"foo": "bar"');
  });

  it('escapes single quotes in curl', () => {
      const toolName = 'quote_tool';
      const args = { text: "I'm testing" };
      const curl = generateCurlCommand(toolName, args);
      // Expected substring with escaping and spacing
      // "text": "I'm testing" -> "text": "I'\''m testing"
      expect(curl).toContain('"text": "I\'\\\'\'m testing"');
  });
});
