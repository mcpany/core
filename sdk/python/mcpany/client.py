# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This module contains the Client class for interacting with an MCP Any server.
"""

from typing import Any, Dict, List, Optional
from contextlib import AsyncExitStack
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from mcp.client.sse import sse_client
from mcp.types import CallToolResult
from .exceptions import ConnectionError, ToolExecutionError

class Client:
    """
    Client for interacting with an MCP Any server.
    """
    def __init__(self, url_or_command: str, args: List[str] = None, token: Optional[str] = None):
        """
        Initialize the MCP Any Client.

        Args:
            url_or_command: The URL of the MCP Any server (e.g., "http://localhost:8080/sse")
                            or the command to run a local server (e.g., "mcpany").
            args: Arguments for the command if running locally.
            token: Bearer token for authentication (only for HTTP/SSE).
        """
        self.url_or_command = url_or_command
        self.args = args or []
        self.token = token
        self._session: Optional[ClientSession] = None
        self._exit_stack: Optional[AsyncExitStack] = None

    async def __aenter__(self):
        await self.connect()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()

    async def connect(self):
        """Connects to the MCP Any server."""
        if self.url_or_command.startswith("http"):
            await self._connect_sse()
        else:
            await self._connect_stdio()

    async def _connect_sse(self):
        self._exit_stack = AsyncExitStack()

        headers = {}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"

        try:
            # We use the sse_client context manager from mcp library
            # It returns (read_stream, write_stream) which we pass to ClientSession
            read_stream, write_stream = await self._exit_stack.enter_async_context(
                sse_client(self.url_or_command, headers=headers)
            )

            self._session = await self._exit_stack.enter_async_context(
                ClientSession(read_stream, write_stream)
            )
            await self._session.initialize()
        except Exception as e:
            await self.close()
            raise ConnectionError(f"Failed to connect to SSE server: {e}") from e

    async def _connect_stdio(self):
        self._exit_stack = AsyncExitStack()

        server_params = StdioServerParameters(
            command=self.url_or_command,
            args=self.args,
            env=None # Inherit env
        )

        try:
            read_stream, write_stream = await self._exit_stack.enter_async_context(
                stdio_client(server_params)
            )

            self._session = await self._exit_stack.enter_async_context(
                ClientSession(read_stream, write_stream)
            )
            await self._session.initialize()
        except Exception as e:
            await self.close()
            raise ConnectionError(f"Failed to connect to Stdio server: {e}") from e

    async def close(self):
        """Closes the connection."""
        if self._exit_stack:
            await self._exit_stack.aclose()
            self._exit_stack = None
            self._session = None

    async def list_tools(self) -> List[Dict[str, Any]]:
        """
        Lists available tools.

        Returns:
            A list of tool definitions (dictionaries).
        """
        if not self._session:
            raise ConnectionError("Client is not connected.")

        result = await self._session.list_tools()
        # Convert ListToolsResult to a simpler list of dicts if needed,
        # but returning the object structure is usually fine.
        # Let's return the list of tools from the result.
        return [tool.model_dump() for tool in result.tools]

    async def call_tool(self, name: str, arguments: Dict[str, Any] = None) -> Any:
        """
        Calls a tool.

        Args:
            name: The name of the tool to call.
            arguments: The arguments for the tool.

        Returns:
            The result of the tool call.
        """
        if not self._session:
            raise ConnectionError("Client is not connected.")

        if arguments is None:
            arguments = {}

        try:
            result: CallToolResult = await self._session.call_tool(name, arguments)

            if result.isError:
                raise ToolExecutionError(f"Tool call failed: {result.content}")

            # Helper to extract text content if simple
            content = result.content
            if len(content) == 1 and content[0].type == "text":
                return content[0].text
            return content

        except Exception as e:
            if isinstance(e, ToolExecutionError):
                raise
            raise ToolExecutionError(f"Failed to call tool {name}: {e}") from e
