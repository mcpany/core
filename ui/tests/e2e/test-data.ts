// Test Data for E2E Tests
export const mockTrace = {
  id: "test-trace-id",
  request_type: "call_tool",
  status: "success",
  duration_ms: 120,
  payload: {
    table: {
      headers: ["id", "name", "email", "role"],
      rows: [
        ["1", "Alice Smith", "alice@example.com", "admin"],
        ["2", "Bob Jones", "bob@example.com", "user"],
      ],
    },
  },
  result: {
    table: {
      headers: ["status", "message"],
      rows: [
        ["success", "Users fetched successfully"],
      ],
    },
  },
  created_at: new Date().toISOString(),
};
