// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements a GraphQL upstream service demo.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
)

func main() {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "RootQuery",
			Fields: graphql.Fields{
				"hello": &graphql.Field{
					Type: graphql.String,
					Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
						return "Hello, world!", nil
					},
				},
			},
		}),
	})
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: r.URL.Query().Get("query"),
		})
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Printf("failed to encode result: %v", err)
		}
	})

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Println("Server is running on port 8080")
	log.Fatal(server.ListenAndServe())
}
