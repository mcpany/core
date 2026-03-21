# File Upload Example

This example demonstrates how to upload a file to the MCP Any server.

## Prerequisites

- [Go](https://golang.org/doc/install) (version 1.24.3 or later)
- [Docker](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/)

## Running the Example

1.  **Start the servers:**

    ```bash
    docker compose up --build
    ```

2.  **Upload a file:**

    ```bash
    curl -X POST -F "file=@dummy.txt" http://localhost:8082/upload
    ```

3.  **Shut down the services:**

    ```bash
    docker compose down
    ```
