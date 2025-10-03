# 🚀 Introduction

Welcome to MCP-X! This document provides a high-level overview of the project, its purpose, and its core components.

MCP-X (Model Context Protocol eXtensions) is a versatile server designed to dynamically register and expose capabilities from various backend services as standardized "Tools." These tools can then be listed and executed through a unified interface. It supports a wide range of upstream services, including gRPC, RESTful APIs (via OpenAPI), generic HTTP services, and command-line tools.

The server also supports authenticating with your upstream services using API keys and bearer tokens, ensuring secure communication with your backends.

The service exposes two main APIs:

- **MCP Router API**: This API allows clients to list and execute the tools that have been registered with the MCP-X server.
- **Registration API**: This API allows backend services to register themselves with the MCP-X server, making their capabilities available as tools.