# Weather.gov API

This directory contains an example of how to use the [weather.gov](https://www.weather.gov/documentation/services-web-api) API.

## Usage

The weather.gov API requires a two-step process to get a weather forecast:

1.  **Get the forecast URL:** Use the `get_grid` call with a latitude and longitude to get a URL for the forecast.
2.  **Get the forecast:** Use the `get_forecast` call with the URL from the previous step to get the weather forecast.

A User-Agent is required to identify your application.

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
./build/bin/server run --config-path examples/popular_services/weather.gov/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

```bash
gemini mcp add --transport http --trust weather.gov http://localhost:50050
```

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "Use get_grid to call get_grid"
```
