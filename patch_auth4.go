package main

import (
    "io/ioutil"
    "log"
    "strings"
)

func main() {
    content, err := ioutil.ReadFile("server/pkg/middleware/auth.go")
    if err != nil {
        log.Fatal(err)
    }

    modified := strings.Replace(string(content), `			// Fallback to method-based extraction if serviceID not yet found
			if serviceID == "" {
				// Extract serviceID from the method. Assuming the format is "service.method".
				// Optimization: Use strings.Cut to avoid allocating a slice.
				if before, _, found := strings.Cut(method, "."); found {
					serviceID = before
				}
			}`, `			// Fallback to method-based extraction if serviceID not yet found
			if serviceID == "" {
				// We don't want to fallback if it's a known service-prefixed parameter method
				// because an attacker could provide an invalid type and bypass auth.
				if method != consts.MethodToolsCall && method != consts.MethodPromptsGet {
					// Extract serviceID from the method. Assuming the format is "service.method".
					// Optimization: Use strings.Cut to avoid allocating a slice.
					if before, _, found := strings.Cut(method, "."); found {
						serviceID = before
					}
				} else {
                    // Sentinel Security Update: Secure By Design, Fail Closed.
					// If we reach here for tools/call or prompts/get, we couldn't extract a valid
                    // serviceID from the payload. This either means the payload is malformed,
                    // or it's an exploit attempt. We MUST fail closed.
                    // By setting a dummy invalid service ID, we force authentication to fail.
                    serviceID = "__invalid_missing_service_id__"
                }
			}`, 1)

    err = ioutil.WriteFile("server/pkg/middleware/auth.go", []byte(modified), 0644)
    if err != nil {
        log.Fatal(err)
    }
}
