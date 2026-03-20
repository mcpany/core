import { unwrapMcpResult, deepParseJson } from './src/lib/mcp-unwrap';

const str = JSON.stringify({
  "content": [
    {
      "type": "text",
      "text": "{\"value\":\"Version 1\"}"
    }
  ],
  "isError": false,
  "value": "Version 1"
});

const res = unwrapMcpResult(str);
console.log("unwrapped:", res);
console.log("deepParsed:", JSON.stringify(deepParseJson(res), null, 2));
