# Popular Services

This directory contains a collection of popular services that have been converted to upstream services for `mcp-any-server`.

## Usage

To use these services, you can load the `config.yaml` files in this directory into your `mcp-any-server` instance. You can do this by adding the following to your `mcp-any-server` configuration:

```bash
mcp-any-server --config-path "examples/popular_services/ipinfo.io"
```

## Contributing

We welcome contributions of new popular services! To contribute, please follow these steps:

1. Create a new subdirectory in this directory with the name of the service (e.g., `ipinfo.io`).
2. In the subdirectory, create a `config.yaml` file that registers the service in `mcp-any-server` format.
3. In the subdirectory, create a `README.md` file that provides an introduction to the service.
4. Add a new row to the table in this `README.md` file with information about the new service.
5. Open a pull request with your changes.

## Services

| Service Name | Introduction                               | Verified Manually | Covered by an Integration Test                         |
| ------------ | ------------------------------------------ | ----------------- | ------------------------------------------------------ |
| ipinfo.io    | Geolocation information for an IP address. | ❌                | [✅](./../tests/integration/popular_service/ipinfo.io) |
