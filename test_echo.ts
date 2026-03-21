import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

const transport = new StdioClientTransport({
  command: "npx",
  args: ["-y", "@modelcontextprotocol/server-everything"]
});

const client = new Client({
  name: "test-client",
  version: "1.0.0"
}, {
  capabilities: {}
});

async function run() {
  await client.connect(transport);
  const tools = await client.listTools();
  console.log(tools.tools.map(t => t.name));
  const res = await client.callTool({
    name: "echo",
    arguments: {
      message: "Hello"
    }
  });
  console.log(res);
  process.exit(0);
}
run();
