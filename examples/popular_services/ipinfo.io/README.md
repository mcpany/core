# ipinfo.io

This service provides geolocation information for an IP address.

## Usage

You can use this service to get your own IP address and location information.

## Authentication

This service does not require authentication for basic use. However, there are rate limits for free usage. You can sign up for an API key to get higher rate limits.
To use an API key, you can add it to the `config.yaml` as follows:

```yaml
upstream_services:
  - name: ipinfo.io
    http_service:
      address: "https://ipinfo.io"
      authentication:
        token: "YOUR_API_KEY"
      calls:
        - endpoint_path: "/json"
          schema:
            name: "ipinfo"
          method: "HTTP_METHOD_GET"
```
