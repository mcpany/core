# Mini-Design: S3 / Blob Storage Adapter

## Goal
Enable `mcpany` to interact with S3-compatible object storage services (AWS S3, MinIO, GCS via S3-compat, etc.) as a first-class upstream service.

## New Configuration
We will add `S3UpstreamService` to `UpstreamServiceConfig`.

```protobuf
message S3UpstreamService {
  // Region (e.g., "us-east-1")
  string region = 1;
  // Bucket name
  string bucket = 2;
  // Endpoint URL (optional, for non-AWS S3)
  string endpoint = 3;
  // Credentials (optional, can fallback to env vars)
  string access_key_id = 4;
  string secret_access_key = 5;
  // Prefix to limit access to (optional)
  string prefix = 6;
  // Read-only mode
  bool read_only = 7;
}
```

## Tools
The adapter will automatically register the following tools:

1.  `list_objects(prefix: string, max_keys: int) -> [ObjectInfo]`
2.  `get_object(key: string) -> content: string`
3.  `put_object(key: string, content: string)` (if not read-only)
4.  `delete_object(key: string)` (if not read-only)
5.  `get_object_metadata(key: string) -> Metadata`

## Implementation Details
-   **Package**: `server/pkg/upstream/s3`
-   **Dependencies**: `github.com/aws/aws-sdk-go-v2/service/s3` and related config packages.
-   **Security**: Ensure paths cannot traverse outside the bucket/prefix (S3 keys are flat, but `prefix` config restricts scope).
-   **Verification**:
    -   Unit tests mocking the S3 client interface.
    -   No live AWS tests in CI, but mock-based verification.

## Edge Cases
-   Large files: `get_object` should probably have a size limit or return a presigned URL (for now, just limit size or read first N bytes).
-   Binary data: `content` in tools is typically text. S3 objects can be binary. We should detect content type or base64 encode if binary. For MVP, we might treat everything as text or base64 based on a flag, or just support text. -> **Decision**: Detect non-utf8 and return base64 encoded string with a flag `is_base64: true`.

## Plan
1.  Modify `proto/config/v1/upstream_service.proto`.
2.  Run `make gen`.
3.  Implement `server/pkg/upstream/s3`.
4.  Register in `server/pkg/upstream/factory`.
