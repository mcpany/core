# Service Registration

MCP Any supports two primary methods for registering upstream services: Static Registration and Dynamic Registration. Both methods allow you to configure service types, authentication, and policies.

## Static Registration

Static registration is the most common way to configure services. It involves defining your services in a YAML or JSON configuration file (e.g., `config.yaml`) that is loaded when the server starts.

### Configuration

You define services under the `upstream_services` key in your configuration file.

```yaml
upstream_services:
  - name: "my-http-service"
    http_service:
      address: "https://api.example.com"
    authentication:
      api_key:
        param_name: "X-Api-Key"
        in: "HEADER"
        key_value: "secret"

  - name: "my-grpc-service"
    grpc_service:
      address: "localhost:50051"
      use_reflection: true
```

### Hot Reloading

If you modify the configuration file while the server is running, MCP Any can automatically detect the changes and reload the configuration without restarting the process (if Hot Reloading is enabled).

## Dynamic Registration

Dynamic registration allows you to add or remove services at runtime using the Admin gRPC API. This is useful for building control planes, UIs, or automated systems that manage MCP Any.

### gRPC API

The Registration Service is exposed via gRPC. You can use any gRPC client to interact with it.

**Service Definition (Proto):**

```protobuf
service RegistrationService {
  rpc RegisterService(RegisterServiceRequest) returns (RegisterServiceResponse);
  rpc ListServices(ListServicesRequest) returns (ListServicesResponse);
  // UnregisterService (Not yet implemented)
}
```

### Example: Registering a Service via gRPC

To register a service, you send a `RegisterServiceRequest` containing the service configuration (same structure as the YAML config).

1.  **Construct the Config Object**: Create a `ServiceConfig` object with the desired service type and settings.
2.  **Call RegisterService**: Send the request to the Admin API.
3.  **Response**: The server will return a `RegisterServiceResponse` containing the `service_key` and a list of discovered tools (if auto-discovery is enabled).

### Benefits of Dynamic Registration

-   **No Restart Required**: Add services without interrupting existing connections.
-   **Automation**: Integrate with service discovery systems (e.g., Consul, K8s) to automatically register new services.
-   **Multi-Tenancy**: Dynamically spin up service adapters for different users or tenants.
