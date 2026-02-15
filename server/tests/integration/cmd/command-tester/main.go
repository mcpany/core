// Package main is a helper command for integration tests.

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
		printEnv      string
	)

	flag.DurationVar(&sleepDuration, "sleep", 0, "sleep duration")
	flag.StringVar(&stdout, "stdout", "", "output to stdout")
	flag.StringVar(&stderr, "stderr", "", "output to stderr")
	flag.IntVar(&exitCode, "exit-code", 0, "exit code")
	flag.StringVar(&printEnv, "print-env", "", "print value of env var")
	flag.Parse()

	if sleepDuration > 0 {
		time.Sleep(sleepDuration)
	}

	if stdout != "" {
		_, _ = fmt.Fprint(os.Stdout, stdout)
	}

	if stderr != "" {
		_, _ = fmt.Fprint(os.Stderr, stderr)
	}

	if printEnv != "" {
		val := os.Getenv(printEnv)
		_, _ = fmt.Fprint(os.Stdout, val)
	}

	os.Exit(exitCode)
}
