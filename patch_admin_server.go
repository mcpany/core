package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	path := "server/pkg/admin/server.go"
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}
	contentStr := string(content)

	if !strings.Contains(contentStr, "func (s *Server) ExportAuditLogs") {
		exportFunc := `

// ExportAuditLogs exports audit logs as CSV.
func (s *Server) ExportAuditLogs(ctx context.Context, req *pb.ListAuditLogsRequest) (*pb.ExportAuditLogsResponse, error) {
	if s.auditMiddleware == nil {
		return nil, status.Error(codes.FailedPrecondition, "audit logging is not enabled")
	}

	var startTime, endTime *time.Time
	if req.GetStartTime() != "" {
		t, err := time.Parse(time.RFC3339, req.GetStartTime())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid start_time: %v", err)
		}
		startTime = &t
	}
	if req.GetEndTime() != "" {
		t, err := time.Parse(time.RFC3339, req.GetEndTime())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid end_time: %v", err)
		}
		endTime = &t
	}

	filter := audit.Filter{
		StartTime: startTime,
		EndTime:   endTime,
		ToolName:  req.GetToolName(),
		UserID:    req.GetUserId(),
		ProfileID: req.GetProfileId(),
	}

	entries, err := s.auditMiddleware.Read(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read audit logs: %v", err)
	}

	var sb strings.Builder
	sb.WriteString("Timestamp,ToolName,UserID,ProfileID,Duration,Status\n")
	for _, e := range entries {
		statusStr := "Success"
		if e.Error != "" {
			statusStr = "Error"
		}
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
			e.Timestamp.Format(time.RFC3339),
			e.ToolName,
			e.UserID,
			e.ProfileID,
			e.Duration,
			statusStr,
		))
	}

	return pb.ExportAuditLogsResponse_builder{CsvData: proto.String(sb.String())}.Build(), nil
}
`
		contentStr += exportFunc
		err = ioutil.WriteFile(path, []byte(contentStr), 0644)
		if err != nil {
			fmt.Println("Error writing:", err)
		} else {
			fmt.Println("Successfully added ExportAuditLogs to server.go")
		}
	} else {
		fmt.Println("ExportAuditLogs already exists in server.go")
	}
}
