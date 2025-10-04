# 🏁 Getting Started

This guide provides a step-by-step walkthrough to get the MCP-XY up and running on your local machine. By following these instructions, you'll be able to build the project, run the server, and verify that everything is working correctly.

## Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.18 or higher)
- [Docker](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/)

## Installation

1. **Clone the repository:**

    ```bash
    git clone https://github.com/mcpxy/core.git
    cd core
    ```

2. **Build the application:**
    This command will generate the necessary protobuf files and build the server executable.

    ```bash
    make build
    ```

## Running the Server

After building the project, you can run the server application:

```bash
make server
```

This will start the MCP-XY server. By default, the server will listen on port `8080`.

You should see log messages indicating that the server has started, for example:

```
INFO main.go:29 Starting MCP-XY...
INFO main.go:51 Attempting to start MCP-XY server on port :8080
INFO main.go:60 MCP-XY server listening on :8080
