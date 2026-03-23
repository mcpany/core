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

    modified := strings.Replace(string(content), `				} else if m, ok := req.(map[string]interface{}); ok {
					if params, ok := m["params"].(map[string]interface{}); ok {
						if name, ok := params["name"].(string); ok {
							if before, _, found := strings.Cut(name, "."); found {
								serviceID = before
							}
						}
					}
				}`, `				}`, 2)

    err = ioutil.WriteFile("server/pkg/middleware/auth.go", []byte(modified), 0644)
    if err != nil {
        log.Fatal(err)
    }
}
