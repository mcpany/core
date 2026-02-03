package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mcpany/core/server/pkg/audit"
)

func main() {
	dbPath := flag.String("db", "/tmp/mcpany.db", "Path to SQLite database")
	flag.Parse()

	log.Printf("Seeding database at %s...", *dbPath)

	store, err := audit.NewSQLiteAuditStore(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open audit store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// 1. Table Data: List of Users
	users := []map[string]any{
		{"id": 1, "name": "Alice Smith", "role": "Admin", "active": true, "last_login": "2023-10-26T10:00:00Z"},
		{"id": 2, "name": "Bob Jones", "role": "User", "active": true, "last_login": "2023-10-25T14:30:00Z"},
		{"id": 3, "name": "Charlie Brown", "role": "User", "active": false, "last_login": "2023-09-12T09:15:00Z"},
		{"id": 4, "name": "Diana Prince", "role": "Editor", "active": true, "last_login": "2023-10-26T11:45:00Z"},
	}
	writeEntry(ctx, store, "list_users", nil, users, "")

	// 2. Markdown Data: A summary
	markdown := `# Weekly Report

## Overview
Performance has **improved** by 15% this week.

### Key Metrics
- **Uptime:** 99.99%
- **Requests:** 1.2M (+5%)
- **Latency:** 45ms (-2ms)

### Incident Log
1. *Mon 10:00 AM*: Minor packet loss.
2. *Tue 02:00 PM*: Database maintenance.

> "Keep up the good work!" - CTO

` + "```go\nfunc main() {\n  fmt.Println(\"Hello\")\n}\n```"

	writeEntry(ctx, store, "generate_report", map[string]any{"type": "weekly"}, markdown, "")

	// 3. Complex JSON: Network Topology
	topology := map[string]any{
		"region": "us-east-1",
		"vpc": map[string]any{
			"id": "vpc-12345",
			"cidr": "10.0.0.0/16",
			"subnets": []map[string]any{
				{"id": "subnet-a", "az": "us-east-1a", "public": true},
				{"id": "subnet-b", "az": "us-east-1b", "public": false},
			},
			"gateways": []string{"igw-123", "ngw-456"},
		},
		"security_groups": []map[string]any{
			{
				"id": "sg-web",
				"rules": []map[string]any{
					{"port": 80, "protocol": "tcp", "source": "0.0.0.0/0"},
					{"port": 443, "protocol": "tcp", "source": "0.0.0.0/0"},
				},
			},
		},
	}
	writeEntry(ctx, store, "get_topology", map[string]any{"region": "us-east-1"}, topology, "")

	// 4. Simple Text
	writeEntry(ctx, store, "echo", map[string]any{"msg": "Hello World"}, "Hello World", "")

	// 5. Error Case
	writeEntry(ctx, store, "deploy_service", map[string]any{"service": "payment"}, nil, "Connection timeout: upstream service unavailable")

	// 6. Large Table (Stress Test)
	largeList := make([]map[string]any, 50)
	for i := 0; i < 50; i++ {
		largeList[i] = map[string]any{
			"index": i,
			"uuid": fmt.Sprintf("uuid-%d", i),
			"value": i * 100,
			"status": func() string {
				if i%3 == 0 {
					return "OK"
				}
				return "PENDING"
			}(),
		}
	}
	writeEntry(ctx, store, "list_transactions", map[string]any{"limit": 50}, largeList, "")

	log.Println("Seeding complete.")
}

func writeEntry(ctx context.Context, store audit.Store, tool string, args any, result any, errStr string) {
	argBytes, _ := json.Marshal(args)
	if args == nil {
		argBytes = []byte("{}")
	}

	entry := audit.Entry{
		Timestamp:  time.Now(),
		ToolName:   tool,
		UserID:     "user-1",
		ProfileID:  "default",
		Arguments:  json.RawMessage(argBytes),
		Result:     result,
		Error:      errStr,
		Duration:   "120ms",
		DurationMs: 120,
	}

	if err := store.Write(ctx, entry); err != nil {
		log.Printf("Failed to write entry for %s: %v", tool, err)
	} else {
		log.Printf("Wrote entry for %s", tool)
	}
}
