import { unwrapMcpResult, deepParseJson } from '../../src/lib/mcp-unwrap';
import { describe, it, expect } from 'vitest';

describe('mcp-unwrap diff', () => {
  it('deepParseJson recursively un-stringifies JSON fields', () => {
    // maybe original is a string itself?
    const original = "{\"isError\":false,\"value\":\"Version 1\",\"content\":[{\"type\":\"text\",\"text\":\"{\\\"value\\\":\\\"Version 1\\\"}\"}]}";
    const output = deepParseJson(unwrapMcpResult(original));
    expect(output.content[0].text.value).toBe("Version 1");
  });

  it('does not over-unwrap strings that should remain strings', () => {
      const output = deepParseJson("Hello");
      expect(output).toBe("Hello");
  });

  it('leaves already parsed json correctly', () => {
      const output = deepParseJson({ text: { value: "Version 1" } });
      expect(output.text.value).toBe("Version 1");
  });
});
