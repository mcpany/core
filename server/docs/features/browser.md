# Browser Provider

The Browser Provider allows the MCP Any server to control a headless Chrome/Chromium browser instance. This enables LLMs to navigate websites, interact with dynamic content (JavaScript), take screenshots, and extract text.

## Configuration

The browser provider is configured as an `upstream_service` with the `browser_service` block.

### Fields

| Field             | Type     | Description                                                                 |
| ----------------- | -------- | --------------------------------------------------------------------------- |
| `endpoint`        | `string` | Optional. The WebSocket Debugger URL of an existing browser (e.g., `ws://localhost:9222`). If empty, a local headless browser is launched. |
| `headless`        | `bool`   | Whether to run in headless mode (no UI). Default `false` (in proto), usually `true` for servers. |
| `viewport_width`  | `int`    | Width of the viewport. Default 1920.                                        |
| `viewport_height` | `int`    | Height of the viewport. Default 1080.                                       |
| `timeout`         | `string` | Timeout for actions (e.g., "30s").                                          |
| `user_agent`      | `string` | Custom User-Agent string.                                                   |

### Example Configuration

```yaml
upstream_services:
  - name: "browser-bot"
    browser_service:
      headless: true
      viewport_width: 1280
      viewport_height: 720
      timeout: "60s"
```

## Tools Exposed

- `navigate(url)`: Go to a URL.
- `screenshot(selector?, full_page?)`: Take a PNG screenshot (returns base64).
- `get_content(selector?, html?)`: Get text or HTML content.
- `click(selector)`: Click an element.
- `type(selector, text)`: Type text into an input.
- `evaluate(expression)`: Run JavaScript code.

## Prerequisites

- **Local Mode**: `google-chrome` or `chromium` must be installed in the environment where `mcpany` runs.
- **Remote Mode**: You can connect to a remote instance (e.g., `browserless.io` or a separate Docker container running Chrome with remote debugging enabled).

## Usage Example

User: "Go to news.ycombinator.com and tell me the top headline."

Model:
1. `browser-bot.navigate(url="https://news.ycombinator.com")`
2. `browser-bot.get_content(selector=".titleline > a")`

Output: "The top headline is 'MCP Any Release v1.0'."
