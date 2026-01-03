# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import pytest
from unittest.mock import AsyncMock, patch, MagicMock
from mcpany.client import Client
from mcpany.exceptions import ConnectionError, ToolExecutionError
from mcp.types import CallToolResult, TextContent, Tool

@pytest.mark.asyncio
async def test_client_connect_sse():
    with patch("mcpany.client.sse_client") as mock_sse, \
         patch("mcpany.client.ClientSession") as mock_session:

        # Mock SSE client context manager
        mock_sse.return_value.__aenter__.return_value = (AsyncMock(), AsyncMock())
        mock_sse.return_value.__aexit__.return_value = None

        # Mock Session context manager
        session_instance = AsyncMock()
        mock_session.return_value.__aenter__.return_value = session_instance
        mock_session.return_value.__aexit__.return_value = None

        client = Client("http://localhost:8080/sse", token="test-token")
        async with client:
            pass

        mock_sse.assert_called_once_with("http://localhost:8080/sse", headers={"Authorization": "Bearer test-token"})
        session_instance.initialize.assert_awaited_once()

@pytest.mark.asyncio
async def test_list_tools():
    with patch("mcpany.client.sse_client") as mock_sse, \
         patch("mcpany.client.ClientSession") as mock_session:

        mock_sse.return_value.__aenter__.return_value = (AsyncMock(), AsyncMock())
        session_instance = AsyncMock()
        mock_session.return_value.__aenter__.return_value = session_instance

        # Mock list_tools response
        mock_tool = Tool(name="test_tool", description="A test tool", inputSchema={})
        session_instance.list_tools.return_value.tools = [mock_tool]

        client = Client("http://localhost:8080/sse")
        await client.connect()
        tools = await client.list_tools()

        assert len(tools) == 1
        assert tools[0]["name"] == "test_tool"
        await client.close()

@pytest.mark.asyncio
async def test_call_tool_success():
    with patch("mcpany.client.sse_client") as mock_sse, \
         patch("mcpany.client.ClientSession") as mock_session:

        mock_sse.return_value.__aenter__.return_value = (AsyncMock(), AsyncMock())
        session_instance = AsyncMock()
        mock_session.return_value.__aenter__.return_value = session_instance

        # Mock call_tool response
        result = CallToolResult(content=[TextContent(type="text", text="success")], isError=False)
        session_instance.call_tool.return_value = result

        client = Client("http://localhost:8080/sse")
        await client.connect()
        res = await client.call_tool("test_tool", {"arg": "val"})

        assert res == "success"
        session_instance.call_tool.assert_awaited_with("test_tool", {"arg": "val"})
        await client.close()

@pytest.mark.asyncio
async def test_call_tool_failure():
    with patch("mcpany.client.sse_client") as mock_sse, \
         patch("mcpany.client.ClientSession") as mock_session:

        mock_sse.return_value.__aenter__.return_value = (AsyncMock(), AsyncMock())
        session_instance = AsyncMock()
        mock_session.return_value.__aenter__.return_value = session_instance

        # Mock call_tool error
        result = CallToolResult(content=[TextContent(type="text", text="error")], isError=True)
        session_instance.call_tool.return_value = result

        client = Client("http://localhost:8080/sse")
        await client.connect()

        with pytest.raises(ToolExecutionError):
            await client.call_tool("test_tool")

        await client.close()
