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

const unwrapped = unwrapMcpResult(result);
console.log(unwrapped);

const prevUnwrapped = deepParseJson(unwrapMcpResult(result));
console.log(prevUnwrapped);
