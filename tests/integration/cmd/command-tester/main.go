// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	var (
		sleepDuration time.Duration
		stdout        string
		stderr        string
		exitCode      int
	)

	flag.DurationVar(&sleepDuration, "sleep", 0, "sleep duration")
	flag.StringVar(&stdout, "stdout", "", "output to stdout")
	flag.StringVar(&stderr, "stderr", "", "output to stderr")
	flag.IntVar(&exitCode, "exit-code", 0, "exit code")
	flag.Parse()

	if sleepDuration > 0 {
		time.Sleep(sleepDuration)
	}

	if stdout != "" {
		fmt.Fprint(os.Stdout, stdout)
	}

	if stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}

	os.Exit(exitCode)
}
