const { unwrapMcpResult, deepParseJson } = require('./ui/src/lib/mcp-unwrap.ts');
console.log(unwrapMcpResult({
  "content": [
    {
      "type": "text",
      "text": "{\"value\":\"Version 1\"}"
    }
  ],
  "isError": false,
  "value": "Version 1"
}));
