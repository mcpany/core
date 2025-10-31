// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
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
