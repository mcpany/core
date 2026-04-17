#!/bin/bash
cat << 'INNER_EOF' > /tmp/lazy_mcp.patch
--- ui/src/components/dashboard/lazy-mcp-dashboard.tsx
+++ ui/src/components/dashboard/lazy-mcp-dashboard.tsx
@@ -9,2 +9,11 @@

+/**
+ * Component that provides a lazy-loaded MCP dashboard for the user.
+ *
+ * @summary Renders the Lazy MCP Dashboard interface.
+ * @param {} props - Component props.
+ * @returns {React.ReactElement} The dashboard component.
+ * @throws {Error} None.
+ * @sideEffects None.
+ */
 export function LazyMcpDashboard() {
INNER_EOF
patch -p0 < /tmp/lazy_mcp.patch
