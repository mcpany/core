import { unwrapMcpResult, deepParseJson } from "./ui/src/lib/mcp-unwrap";

const result = {
  "content": [
    {
      "type": "text",
      "text": "{\"value\":\"Version 1\"}"
    }
  ],
  "isError": false,
  "value": "Version 1"
};

console.log("unwrap:", JSON.stringify(unwrapMcpResult(result), null, 2));
console.log("deepParse:", JSON.stringify(deepParseJson(unwrapMcpResult(result)), null, 2));
