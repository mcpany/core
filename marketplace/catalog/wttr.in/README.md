# wttr.in API

This directory contains an example of how to use the [wttr.in](https://wttr.in) API.

## Tools

### `get_weather`

Gets the weather forecast for a specified location in structured JSON format.

**Parameters:**

- `location`: The location to query (City, Airport, GPS, etc.)
- `lang`: Language code (optional, defaults to `en`)

### `get_moon_phase`

Gets the Moon phase information. Returns text/ASCII art.

**Parameters:**

- `args`: Optional arguments to specify date or location context (defaults to empty string).
  - Date: Append `@YYYY-MM-DD` (e.g., `@2023-12-25`)
  - Location: Append `,Location` (e.g., `,London`)

## Usage

```bash
# From the root of the repository
make build
./build/bin/server run --config-path ./examples/popular_services/wttr.in/config.yaml
```

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
./build/bin/server run --config-path examples/popular_services/wttr.in/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust wttr.in http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "Use get_weather to Get the weather forecast for a location in JSON format."
```
