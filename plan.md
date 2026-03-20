Wait, the backend API endpoints were failing `ECONNREFUSED` because the backend wasn't running during the Playwright test!
Playwright test needs the backend to be running if it's hitting `/api/v1/debug/seed`!
And `make test` failed in `make test-proto` because of some broken proto generated files?
Wait, `make test-proto` generates the proto files and compiles them. But they have errors: `undefined: AdminServiceClient`.
Let me run `make clean` and `make build`.
