# Copyright (C) 2025 Author(s) of MCP-XY
# SPDX-License-Identifier: Apache-2.0

import asyncio
import json
import sys

import cowsay
from fastmcp import mcp


async def main():
    @mcp.tool()
    def say(message: str) -> str:
        """Says a message using cowsay."""
        return cowsay.get_output_string("cow", message)

    await mcp.run_stdio()


if __name__ == "__main__":
    if "--mcp-stdio" in sys.argv:
        asyncio.run(main())
    else:
        print("This is a mock MCP service and should be run with --mcp-stdio", file=sys.stderr)
        sys.exit(1)
