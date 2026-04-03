# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import asyncio
import json
import sys

import cowsay
from fastmcp import FastMCP

mcp = FastMCP("e2e-cowsay-server")

def main():
    """
    main operation.

    Summary: Main the  appropriately based on current system conditions.

    Parameters:
      - Check signature.

    Returns:
      - Expected state or data.

    Throws/Errors:
      - Propagates errors.

    Side Effects:
      - Depends on implementation.
    """
    @mcp.tool()
    def say(message: str) -> str:
        """Says a message using cowsay."""
        return cowsay.get_output_string("cow", message)

    mcp.run(transport='stdio', show_banner=False)


if __name__ == "__main__":
    main()
