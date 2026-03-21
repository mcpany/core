# MCPANY Examples

This directory contains examples of how to use MCPANY with different upstream service types. Each example includes:

- An upstream service implementation.
- A configuration file for MCPANY.
- A shell script to start an MCPANY server configured for the example.
- A README file with instructions on how to run the example and interact with it using AI tools like Gemini CLI.

## 1. Build the MCPANY binary

Before running any of the examples, you need to build the `mcpany` binary. From the root of the `core` project, run:

```bash
make build
```

This will create the `mcpany` binary in the `build/bin` directory. All example scripts use this path.

## 2. Creating a Simple MCP Server (Time, Location, and Weather Example)

This example demonstrates how to expose multiple public APIs as tools through `mcpany`.

### Step 1: The `mcpany` Configuration

The `mcpany` configuration file (`upstream_service/http/config/mcpany_config.yaml`) tells `mcpany` how to connect to the upstream public APIs and what tools to expose. Here's what it looks like:

```yaml
global_settings:
  bind_address: "0.0.0.0:8080"
upstream_services:
  - name: ip-location-service
    http_service:
      address: "https://ipapi.co"
      calls:
        - operation_id: "getLocation"
          description: "Gets the user's current location based on their IP address."
          endpoint_path: "/json/"
          method: "HTTP_METHOD_GET"
  - name: weather-service
    http_service:
      address: "https://api.open-meteo.com"
      calls:
        - operation_id: "getWeather"
          description: "Gets the current weather for a given latitude and longitude."
          endpoint_path: "/v1/forecast"
          method: "HTTP_METHOD_GET"
  - name: time-service
    http_service:
      address: "http://localhost:8081"
      calls:
        - operation_id: "GET/time"
          endpoint_path: "/time"
          method: "HTTP_METHOD_GET"
```

This configuration defines three tools:

- `getLocation`: Gets the user's location based on their IP address using the `ipapi.co` service.
- `getWeather`: Gets the current weather for a given latitude and longitude using the `Open-Meteo` API.
- `GET/time`: Gets the current time from a local Go server.

### Step 2: Running the Example

1. **Run the `mcpany` Server**

   In a terminal, start the `mcpany` server using the provided shell script from the `upstream_service/http` directory:

   ```bash
   ./start.sh
   ```

   The `mcpany` server will start and listen for JSON-RPC requests on port `8080`.

   _(Note: The local time server is not required for the location and weather tools to function.)_

### Step 3: Interacting with the Tools (Chained Example)

This example showcases how an AI can chain tools together to perform a more complex task.

#### Using Gemini CLI

1. **Add the MCP Server to Gemini CLI:**

   Use the `gemini mcp add` command to register the running `mcpany` server.

   ```bash
   gemini mcp add mcpany-http-example --transport http http://localhost:8080
   ```

2. **List Available Tools:** Use the `gemini list tools` command to see the tools exposed by the `mcpany` server. You should see `ip-location-service-df37f29a/getLocation`, `weather-service-somehash/getWeather`, and `time-service-c0eda60c/get_time_by_ip`.

3. **Call the `getLocation` Tool:** First, get the location for a given IP address.

   ```bash
   gemini -m gemini-2.5-flash -p "What is the location of the IP address 8.8.8.8?"
   ```

   You should receive a JSON response with the location information, including `latitude` and `longitude`.

4. **Call the `getWeather` Tool:** Now, use the latitude and longitude from the previous step to get the weather.

   ```bash
   # Replace with the actual latitude and longitude from the previous step
   gemini -m gemini-2.5-flash -p "What is the weather at latitude YOUR_LATITUDE and longitude YOUR_LONGITUDE?"
   ```

   This demonstrates how an AI could first determine the user's location and then use that information to provide a local weather forecast.

## 3. Running Other Examples

Each example has a `start.sh` script that starts the MCPANY server with the correct configuration. The upstream service needs to be started separately. Please refer to the `README.md` file in each example's directory for detailed instructions.
