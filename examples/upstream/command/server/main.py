# Copyright (C) 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import asyncio
import sys
from fastmcp import mcp

@mcp.tool()
async def hello(name: str) -> str:
    """A simple hello world command."""
    return f"Hello, {name}!"

async def main():
    # The tool is registered via the decorator, so we just need to run the server.
    await mcp.run_stdio()

if __name__ == "__main__":
    if "--mcp-stdio" in sys.argv:
        asyncio.run(main())
    else:
        print("This script is an MCP service and must be run with --mcp-stdio", file=sys.stderr)
        sys.exit(1)
