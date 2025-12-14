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
