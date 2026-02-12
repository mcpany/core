// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main is the entrypoint for the connector runtime.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// main is the entry point for the connector runtime.
//
// This runtime is designed to host MCP connectors (stdin/stdout tools)
// as a sidecar process, managing their lifecycle and configuration.
// It delegates execution to the run function and handles exit codes.
func main() {
	if err := run(os.Args, make(chan os.Signal, 1)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the connector runtime logic.
//
// It parses command-line arguments, starts the runtime in the specified mode
// (e.g., sidecar), and waits for termination signals.
//
// Parameters:
//   - args: []string. The command-line arguments.
//   - stop: chan os.Signal. A channel to receive termination signals.
//
// Returns:
//   - error: An error if argument parsing or runtime initialization fails.
func run(args []string, stop chan os.Signal) error {
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	connectorName := fs.String("name", "", "Name of the connector to run")
	sidecar := fs.Bool("sidecar", true, "Run in sidecar mode")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if *connectorName == "" {
		return fmt.Errorf("usage: connector-runtime -name <connector> [-sidecar]")
	}

	log.Printf("Starting Connector Runtime for: %s", *connectorName)
	if *sidecar {
		log.Printf("Mode: Sidecar")
	}

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Runtime active. Waiting for signals...")

	<-stop
	log.Println("Shutting down connector runtime...")
	return nil
}
