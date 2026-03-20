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

console.log("deepParse raw:", JSON.stringify(deepParseJson(result), null, 2));
