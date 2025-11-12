# NASA Global Imagery Browse Services (GIBS)

This service provides access to global, full-resolution satellite imagery from NASA's Global Imagery Browse Services (GIBS).

## Usage

You can use this service to retrieve a map tile for a specific location, date, and imagery layer.

To get a map tile, call the `get_tile` tool with the following parameters:

- `LayerIdentifier`: The name of the imagery layer.
- `Time`: The date of the imagery (YYYY-MM-DD).
- `TileMatrixSet`: The tile matrix set (e.g., '250m').
- `TileMatrix`: The zoom level.
- `TileRow`: The row number of the tile.
- `TileCol`: The column number of the tile.

Here is an example of how to call the `get_tile` tool using `curl`:

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "nasa-gibs/-/get_tile", "arguments": {"LayerIdentifier": "MODIS_Terra_CorrectedReflectance_TrueColor", "Time": "2012-07-09", "TileMatrixSet": "250m", "TileMatrix": "6", "TileRow": "13", "TileCol": "36"}}, "id": 1}' \
  http://localhost:50050
```

## Authentication

This service does not require authentication.
