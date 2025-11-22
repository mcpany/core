# Gemini

This example demonstrates how to use `mcpany` to connect to the Gemini API.

## Prerequisites

- An API key for the Gemini API. You can obtain one from [Google AI Studio](https://makersuite.google.com/).

## Configuration

The configuration for this example is in `config.yaml`. It defines a single upstream service for Gemini.

To use this example, you need to set the `GEMINI_API_KEY` environment variable to your Gemini API key.

## Usage

1. **Set the `GEMINI_API_KEY` environment variable:**

   ```bash
   export GEMINI_API_KEY="YOUR_GEMINI_API_KEY"
   ```

2. **Run `mcpany` with the Gemini configuration:**

   ```bash
   make run ARGS="--config-paths ./examples/popular_services/gemini/config.yaml"
   ```

3. **Call the `generateContent` tool:**

   ```bash
   curl -X POST -H "Content-Type: application/json" \
     -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "gemini/-/generateContent", "arguments": {"contents": [{"parts": [{"text": "Write a story about a magic backpack."}]}]}}, "id": 1}' \
     http://localhost:50050
   ```
