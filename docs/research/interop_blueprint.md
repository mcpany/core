# Interop Blueprint: Universal Agent Bus

## Overview
This document serves as the foundation for how the MCP Any Universal Adapter implements and supports a wider range of agentic standard protocols.

## Supported Protocols
- **MCP (Model Context Protocol)**: Supported as a baseline for tool interactions and integrations.
- **ACP (Agent Control Protocol)**: Adds control signals allowing direct management, state injection, and manipulation across complex tasks.
- **A2A (Agent-to-Agent)**: Supported for direct communication and task offloading between autonomous swarms.

## Supported Frameworks & Adapters
1. **OpenClaw**
   - Protocols: `MCP`, `ACP`
   - Description: Supports baseline interactions and comprehensive control capabilities.
2. **CrewAI**
   - Protocols: `MCP`, `A2A`
   - Description: Built for dynamic swarms. Translates multi-agent tasks natively over A2A boundaries.
3. **AutoGen**
   - Protocols: `MCP`, `A2A`, `ACP`
   - Description: Advanced general-purpose multi-agent framework fully leveraging all provided interop protocols.

## Interop Request Flow
All frameworks communicate seamlessly over the interop bus via `InteropRequest`, which is evaluated against available adapter capabilities. Protocol mismatches are actively blocked, returning distinct compatibility errors.
