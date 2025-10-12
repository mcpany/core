# FastMcp Upstream Example

This example shows how to expose a python command as a tool through MCPXY using fastmcp.

## 1. Build the MCPXY binary

From the root of the `core` project, run:

```bash
make build
```

This will create the `mcpxy` binary in the `bin` directory.

## 2. Make the script executable

Before running the MCPXY server, you need to make the `hello.sh` script executable. From the `server` directory, run:

```bash
chmod +x hello.sh
```

## 3. Run the MCPXY Server

In a terminal, start the MCPXY server which is configured to expose the command. From the root of the `fastmcp` example directory, run:

```bash
./start.sh
```

The MCPXY server will start on port 8080.

## 4. Interact with the Tool using Gemini CLI

Now you can use an AI tool like Gemini CLI to interact with the command through MCPXY.

### Configuration

First, configure Gemini CLI to use the local MCPXY server as a tool extension.
```bash
gemini mcp add mcpxy-fastmcp-hello http://localhost:8080
```

### List Available Tools

Now, you can ask Gemini CLI to list the available tools:

```
$ gemini list tools
```

You should see the `hello-service.hello` tool in the list.

### Test the Service

Finally, you can test the service by asking Gemini CLI to call the tool:

```
$ gemini call tool hello-service.hello -- name friend
```

You should see a response similar to this:

```
Hello, friend!
```