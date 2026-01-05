# Hot Reload

**Status**: Implemented

Hot Reload allows the MCP Any server to detect changes to its configuration files and automatically reload them without requiring a full process restart. This significantly improves the developer experience by providing immediate feedback.

## How it works

The server uses a file watcher (based on `fsnotify`) to monitor the configuration files specified at startup. When a write event is detected, the server debounces the event and triggers a reload sequence.

## Limitations

*   Currently supports reloading of the main configuration file and referenced files.
*   Some stateful connections might be reset during reload.
