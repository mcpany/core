# Mock Manifest

## UI Mocks
- `ui/src/lib/marketplace-service.ts`: Contains `MOCK_OFFICIAL_COLLECTIONS` and `fetchOfficialCollections`.
- `ui/src/lib/marketplace-data.ts`: Contains `MARKETPLACE_ITEMS` (unused/dead code).
- `ui/src/mocks/proto/mock-proto.ts`: Generated protobuf mocks.

## Server Mocks
- `server/pkg/app/seed.go`: Contains `SeedRequest` and `handleDebugSeed` (used for seeding, legitimate test util).
