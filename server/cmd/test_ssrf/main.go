package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/mcpany/core/server/pkg/util"
)

func main() {
	// Start a local server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	fmt.Printf("Listening on %s\n", addr)

	go func() {
		http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	}()

	// Try to connect using NewSafeHTTPClient
	client := util.NewSafeHTTPClient()
	_, err = client.Get("http://" + addr)
	if err != nil {
		fmt.Printf("NewSafeHTTPClient: FAILED (err=%v)\n", err)
	} else {
		fmt.Printf("NewSafeHTTPClient: SUCCESS\n")
	}

	// Try to connect using CheckConnection
	err = util.CheckConnection(context.Background(), addr)
	if err != nil {
		fmt.Printf("CheckConnection: FAILED (err=%v)\n", err)
	} else {
		fmt.Printf("CheckConnection: SUCCESS\n")
	}
}
