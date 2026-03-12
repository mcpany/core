//go:build ignore

package main

import (
"encoding/json"
"fmt"
configv1 "github.com/mcpany/core/proto/config/v1"
"google.golang.org/protobuf/encoding/protojson"
)

func main() {
testServices := []string{
`{"id":"svc_01","name":"Payment Gateway","version":"v1.2.0","http_service":{"address":"https://stripe.com","tools":[{"name":"process_payment","description":"Process a payment","call_id":"process_payment_call"}],"calls":{"process_payment_call":{"method":"HTTP_METHOD_POST","endpoint_path":"/v1/charges"}}}}`,
`{"id":"svc_02","name":"User Service","version":"v1.0","http_service":{"address":"http://localhost:50051","tools":[{"name":"get_user","description":"Get user details","call_id":"get_user_call"}],"calls":{"get_user_call":{"method":"HTTP_METHOD_GET","endpoint_path":"/users/{id}"}}}}`,
`{"id":"svc_03","name":"Math","version":"v1.0","http_service":{"address":"http://localhost:5678","tools":[{"name":"calculator","description":"calc","call_id":"calc_call"}],"prompts":[{"name":"Calculate Sum","description":"Adds two numbers together","inputSchema":{"type":"object","properties":{"a":{"type":"number","description":"First number"},"b":{"type":"number","description":"Second number"}},"required":["a","b"]}}],"calls":{"calc_call":{"method":"HTTP_METHOD_POST","endpoint_path":"/calc"}}}}`,
`{"id":"svc_echo","name":"Echo Service","version":"v1.0","command_line_service":{"command":"echo","tools":[{"name":"echo_tool","description":"Echoes back input","input_schema":{"type":"object"},"call_id":"echo_call"}],"calls":{"echo_call":{"args":["echoed_output"]}}}}`,
}

for i, svcJSON := range testServices {
s := configv1.UpstreamServiceConfig_builder{}.Build()
if err := protojson.Unmarshal([]byte(svcJSON), s); err != nil {
fmt.Printf("Service %d error: %v\n", i, err)
} else {
fmt.Printf("Service %d OK: name=%s\n", i, s.GetName())
}
}

// Test template with camelCase oauth2 fields
templateJSON := `{"id":"google-calendar","name":"Google Calendar","description":"Manage events.","icon":"calendar","tags":["google"],"service_config":{"name":"google_calendar","upstream_auth":{"oauth2":{"token_url":"https://oauth2.googleapis.com/token","client_id":{"plainText":""},"client_secret":{"plainText":""},"scopes":"https://www.googleapis.com/auth/calendar"}},"openapi_service":{"spec_url":"https://example.com/spec.yaml"}}}`
t := configv1.ServiceTemplate_builder{}.Build()
if err := protojson.Unmarshal([]byte(templateJSON), t); err != nil {
fmt.Printf("Template error: %v\n", err)
} else {
fmt.Printf("Template OK: id=%s\n", t.GetId())
}

userJSON := `{"id":"e2e-admin-core","authentication":{"basic_auth":{"username":"e2e-admin-core","password_hash":"$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a"}},"roles":["admin"],"profile_ids":["dev","prod"]}`
u := configv1.User_builder{}.Build()
if err := protojson.Unmarshal([]byte(userJSON), u); err != nil {
fmt.Printf("User error: %v\n", err)
} else {
fmt.Printf("User OK: id=%s\n", u.GetId())
}

_ = json.Marshal
}
