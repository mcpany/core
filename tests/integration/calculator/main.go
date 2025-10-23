package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
)

type numbers struct {
	A int `json:"a"`
	B int `json:"b"`
}

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	handler := http.NewServeMux()
	handler.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		var n numbers
		err := json.NewDecoder(r.Body).Decode(&n)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, `{"result": %d}`, n.A+n.B)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: handler,
	}

	fmt.Printf("Calculator server listening on port %d\n", *port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Printf("Calculator server ListenAndServe error: %v\n", err)
	}
}
