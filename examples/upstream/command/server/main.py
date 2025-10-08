import asyncio
from fastmcp import start_mcp_server, tool

@tool
async def hello(name: str) -> str:
    """A simple hello world command."""
    return f"Hello, {name}!"

if __name__ == "__main__":
    asyncio.run(start_mcp_server(tools=[hello]))