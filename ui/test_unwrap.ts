import { unwrapMcpResult, deepParseJson } from './src/lib/mcp-unwrap';

const res = unwrapMcpResult({
  "content": [
    {
      "type": "text",
      "text": "{\"value\":\"Version 1\"}"
    }
  ],
  "isError": false,
  "value": "Version 1"
});

console.log("unwrapped:", res);
console.log("deepParsed:", deepParseJson(res));
