#!/usr/bin/env python3
# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


from datetime import datetime
import sys

from fastmcp import start_mcp_server

def get_local_time_json(args):
    """
    get_local_time_json operation.

    Summary: Fetches and returns the underlying local time json from the system state.

    Parameters:
      - Check signature.

    Returns:
      - Expected state or data.

    Throws/Errors:
      - Propagates errors.

    Side Effects:
      - Depends on implementation.
    """
    now = datetime.now().astimezone()
    time_str = now.strftime("%Y-%m-%d %H:%M:%S")
    timezone_str = now.tzname()
    return {
        "current_time": time_str,
        "timezone": timezone_str
    }

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
    start_mcp_server(
        service_name="time-server",
        service_description="A simple service that provides the current time.",
        tools={"get_local_time": get_local_time_json},
        input_stream=sys.stdin,
        output_stream=sys.stdout
    )

if __name__ == "__main__":
    main()
