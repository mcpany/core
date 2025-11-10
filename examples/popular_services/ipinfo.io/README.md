# ipinfo.io

This service provides geolocation information for an IP address.

## Usage

You can use this service to get your own IP address and location information, or to look up a specific IP address.

To get your own IP address information, simply call the `ipinfo` tool with no parameters:

```
mcp call ipinfo_io.ipinfo
```

To look up a specific IP address, provide the `ip` parameter:

```
mcp call ipinfo_io.ipinfo --ip 8.8.8.8
```

## Authentication

This service does not require authentication for basic use. However, there are rate limits for free usage. You can sign up for an API key to get higher rate limits.

To use an API key, set the `IPINFO_API_TOKEN` environment variable to your API key. The `config.yaml` file is already configured to use this environment variable.
