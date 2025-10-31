# 🏁 Getting Started

This guide provides a step-by-step walkthrough to get the MCP Any up and running on your local machine. By following these instructions, you'll be able to build the project, run the server, and verify that everything is working correctly.

## Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.24.3 or higher)
- [Docker](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/)

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/mcpany/core.git
   cd core
   ```

2. **Build the application:**
   This command will generate the necessary protobuf files and build the server executable into `./build/bin/server`.

   ```bash
   make build
   ```

## Running the Server

After building the project, you can run the server application:

```bash
make run
```

This will start the MCP Any server. By default, the server will listen for JSON-RPC requests on port `50050`.

You should see log messages indicating that the server has started, for example:

```
INFO Starting MCP Any server locally... service=mcpany
INFO Configuration jsonrpc-port=50050 registration-port= grpc-port= stdio=false config-paths=[] service=mcpany
INFO Application started service=mcpany
```
