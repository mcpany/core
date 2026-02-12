# ipinfo.io

This service provides geolocation information for an IP address.

## Usage

You can use this service to get your own IP address and location information, or to look up a specific IP address.

To get your own IP address information, simply call the `ipinfo` tool with no parameters:

```
gemini -m gemini-2.5-flash -p 'what is my current ip information'
```

To look up a specific IP address, provide the `ip` parameter:

```
gemini -m gemini-2.5-flash -p 'what is ip information for IP 8.8.8.8'
```

## Authentication

This service does not require authentication for basic use. However, there are rate limits for free usage. You can sign up for an API key to get higher rate limits.

To use an API key, set the `IPINFO_API_TOKEN` environment variable to your API key. The `config.yaml` file is already configured to use this environment variable.

## Usage with Gemini CLI

### 1. Start the MCP Server

```bash
# From repo root
make build # if not already built
# Export required environment variables
export IPINFO_API_TOKEN=YOUR_IPINFO_API_TOKEN_VALUE

./build/bin/server run --config-path examples/popular_services/ipinfo.io/config.yaml
```

### 2. Add to Gemini

In a separate terminal:

Access the service at your server's address (default `http://localhost:50050` or as configured).

### 3. Example Query

```bash
gemini -m gemini-2.5-flash -p "Use ipinfo to call ipinfo"
```
