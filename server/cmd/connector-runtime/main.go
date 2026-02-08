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
// Summary: Executes the connector runtime.
//
// Side Effects:
//   - Starts the runtime process.
//   - Listens for signals.
//   - Exits with status 1 on error.
func main() {
	if err := run(os.Args, make(chan os.Signal, 1)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

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
