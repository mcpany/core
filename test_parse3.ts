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
console.log("unwrap returns:", unwrapped);
console.log("isArray?", Array.isArray(unwrapped));

const fullyParsed = deepParseJson(unwrapped);
console.log("deepParse returns:", fullyParsed);
