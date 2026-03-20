// If it fails with 500:
// "Failed to seed trace" -> means standardMiddlewares.Audit.Write returned error
// Let's check `api_traces.go` again:
//
// 			if err := a.standardMiddlewares.Audit.Write(r.Context(), entry); err != nil {
// 				logging.GetLogger().Error("failed to seed trace to audit db", "error", err)
// 				http.Error(w, "Failed to seed trace", http.StatusInternalServerError)
// 				return
// 			}
//
// What error could `Write` return?
// Write does:
// 	resultJSON := "{}"
// 	if entry.Result != nil {
// 		if b, err := json.Marshal(entry.Result); err == nil {
// 			resultJSON = string(b)
// 		}
// 	}
// Wait, `Write` also executes SQL: `INSERT INTO audit_logs ...`
// Does it fail to insert?
//
