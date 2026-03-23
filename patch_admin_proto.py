with open("proto/admin/v1/admin.proto", "r") as f:
    content = f.read()

content = content.replace("""  // ListAuditLogs returns audit logs matching the filter.
  rpc ListAuditLogs(ListAuditLogsRequest) returns (ListAuditLogsResponse) {
    option (google.api.http) = {
      get: "/api/v1/audit/logs"
    };
  }
}""", """  // ListAuditLogs returns audit logs matching the filter.
  rpc ListAuditLogs(ListAuditLogsRequest) returns (ListAuditLogsResponse) {
    option (google.api.http) = {
      get: "/api/v1/audit/logs"
    };
  }

  // ExportAuditLogs exports audit logs as CSV.
  rpc ExportAuditLogs(ListAuditLogsRequest) returns (ExportAuditLogsResponse) {
    option (google.api.http) = {
      get: "/api/v1/audit/export"
    };
  }
}""")

if "message ExportAuditLogsResponse" not in content:
    content = content.replace("// AuditLogEntry", """// ExportAuditLogsResponse contains the exported CSV data.
message ExportAuditLogsResponse {
  // The exported CSV data.
  string csv_data = 1;
}

// AuditLogEntry""")

with open("proto/admin/v1/admin.proto", "w") as f:
    f.write(content)
