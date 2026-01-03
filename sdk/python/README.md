# MCP Any Python Client SDK

A Python client library for connecting to [MCP Any](https://github.com/mcpany/core) servers.

## Installation

```bash
pip install mcpany
```

## Usage

### Connecting via SSE (HTTP)

```python
import asyncio
from mcpany import Client

async def main():
    # Connect to a running MCP Any server
    async with Client("http://localhost:8080/sse", token="my-api-key") as client:
        # List available tools
        tools = await client.list_tools()
        print(f"Available tools: {[t['name'] for t in tools]}")

        # Call a tool
        try:
            result = await client.call_tool("weather_get_forecast", {"city": "London"})
            print(f"Result: {result}")
        except Exception as e:
            print(f"Error: {e}")

if __name__ == "__main__":
    asyncio.run(main())
```

### Connecting via Stdio (Local Command)

```python
import asyncio
from mcpany import Client

async def main():
    # Run a local MCP server command
    async with Client("mcpany", args=["run", "--config", "config.yaml"]) as client:
        tools = await client.list_tools()
        print(tools)

if __name__ == "__main__":
    asyncio.run(main())
```

## Development

1. Install dependencies:
   ```bash
   pip install -e .[test]
   ```

2. Run tests:
   ```bash
   pytest
   ```
