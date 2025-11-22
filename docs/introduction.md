# 🚀 Introduction

Welcome to MCP Any! This document provides a high-level overview of the project, its purpose, and its core components.

## The Goal: Elimination of Custom MCP Servers

The primary goal of **MCP Any** is to eliminate the requirement to implement, compile, and maintain new MCP servers for every API call you want to expose.

Traditionally, if you wanted to expose an API as an MCP tool, you would need to write a specific server wrapper for it. With MCP Any, you can achieve this using **pure configuration**.

## Core Philosophy

1.  **Configuration-First**: Users can create MCP server tools, resources, and prompts with just configuration files (YAML/JSON).
2.  **Shareability**: Configurations can be easily shared publicly. This means the community can exchange "connectors" for various services without passing around opaque binaries or requiring complex build environments.
3.  **Local & Secure**: By running MCP Any locally, you keep your sensitive API keys and data within your own environment. You don't risk leaking information to a remote proxy service just to convert an API call to MCP.

## How It Works

MCP Any acts as a versatile server that interprets your configurations and dynamically "becomes" the MCP server you need. It supports a wide range of upstream services, including gRPC, RESTful APIs (via OpenAPI), generic HTTP services, and command-line tools.

The service exposes two main APIs:

- **MCP Router API**: This API allows clients to list and execute the tools that have been registered with the MCP Any server.
- **Registration API**: This API allows backend services to register themselves with the MCP Any server, making their capabilities available as tools.
