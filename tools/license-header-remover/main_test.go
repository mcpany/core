package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsSourceFile(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Go file", "main.go", true},
		{"Python file", "script.py", true},
		{"Shell script", "run.sh", true},
		{"YAML file", "config.yaml", true},
		{"YML file", "config.yml", true},
		{"Proto file", "service.proto", true},
		{"Makefile", "Makefile", true},
		{"Dockerfile", "Dockerfile", true},
		{"Text file", "notes.txt", false},
		{"Image file", "logo.png", false},
		{"No extension", "binary", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSourceFile(tc.path); got != tc.expected {
				t.Errorf("isSourceFile(%q) = %v, want %v", tc.path, got, tc.expected)
			}
		})
	}
}

const dummyGoFileContent = `// Copyright 2025
package main
`

func TestMain(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy go file with a license header
	goFileContent := dummyGoFileContent
	goFilePath := filepath.Join(tmpDir, "main.go")
	// Create dummy Go file
	if err := os.WriteFile(goFilePath, []byte(goFileContent), 0600); err != nil {
		t.Fatalf("Failed to create dummy Go file: %v", err)
	}

	// Create dummy Python file
	pyFileContent := `
# Copyright 2025
# Header
print("Hello")
`
	pyFilePath := filepath.Join(tmpDir, "script.py")
	if err := os.WriteFile(pyFilePath, []byte(pyFileContent), 0600); err != nil {
		t.Fatalf("Failed to create dummy Python file: %v", err)
	}

	// Create a subdirectory and a file in it
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0750); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	subFileContent := dummyGoFileContent
	subFilePath := filepath.Join(subDir, "main.go")
	if err := os.WriteFile(subFilePath, []byte(subFileContent), 0600); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	// Create a .git directory to be skipped
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0750); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create a vendor directory to be skipped
	vendorDir := filepath.Join(tmpDir, "vendor")
	if err := os.Mkdir(vendorDir, 0750); err != nil {
		t.Fatalf("Failed to create vendor directory: %v", err)
	}

	// Create a build directory to be skipped
	buildDir := filepath.Join(tmpDir, "build")
	if err := os.Mkdir(buildDir, 0750); err != nil {
		t.Fatalf("Failed to create build directory: %v", err)
	}

	// Create a node_modules directory to be skipped
	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	if err := os.Mkdir(nodeModulesDir, 0750); err != nil {
		t.Fatalf("Failed to create node_modules directory: %v", err)
	}

	// Create a .pb.go file to be skipped
	pbgoFileContent := dummyGoFileContent
	pbgoFilePath := filepath.Join(tmpDir, "main.pb.go")
	if err := os.WriteFile(pbgoFilePath, []byte(pbgoFileContent), 0600); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	// Change the current working directory to the temporary directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	// Call the main function
	os.Args = []string{"license-header-remover", "."}
	main()

	// Check if the license header has been removed from the go file
	goFileContentAfter, err := os.ReadFile(goFilePath)
	if err != nil {
		t.Fatalf("Failed to read go file: %v", err)
	}
	if strings.Contains(string(goFileContentAfter), "Copyright") {
		t.Errorf("License header not removed from go file")
	}

	// Check if the license header has been removed from the python file
	pyFileContentAfter, err := os.ReadFile(pyFilePath)
	if err != nil {
		t.Fatalf("Failed to read python file: %v", err)
	}
	if strings.Contains(string(pyFileContentAfter), "Copyright") {
		t.Errorf("License header not removed from python file")
	}

	// Check if the license header has been removed from the file in the subdirectory
	subFileContentAfter, err := os.ReadFile(subFilePath)
	if err != nil {
		t.Fatalf("Failed to read sub file: %v", err)
	}
	if strings.Contains(string(subFileContentAfter), "Copyright") {
		t.Errorf("License header not removed from sub file")
	}

	// Check if the .pb.go file is untouched
	pbgoFileContentAfter, err := os.ReadFile(pbgoFilePath)
	if err != nil {
		t.Fatalf("Failed to read .pb.go file: %v", err)
	}
	if !strings.Contains(string(pbgoFileContentAfter), "Copyright") {
		t.Errorf("License header removed from .pb.go file")
	}

}

func TestFindBlockComment(t *testing.T) {
	testCases := []struct {
		name          string
		lines         []string
		startIdx      int
		expectedStart int
		expectedEnd   int
	}{
		{
			"Simple case",
			[]string{"/*", "Copyright", "*/"},
			1, 0, 2,
		},
		{
			"No start",
			[]string{"Copyright", "*/"},
			0, -1, -1,
		},
		{
			"No end",
			[]string{"/*", "Copyright"},
			1, -1, -1,
		},
		{
			"Offset start",
			[]string{"", "/*", "Copyright", "*/"},
			2, 1, 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start, end := findBlockComment(tc.lines, tc.startIdx)
			if start != tc.expectedStart || end != tc.expectedEnd {
				t.Errorf("findBlockComment() = (%v, %v), want (%v, %v)", start, end, tc.expectedStart, tc.expectedEnd)
			}
		})
	}
}

func TestRefineEndIndex(t *testing.T) {
	testCases := []struct {
		name     string
		lines    []string
		start    int
		end      int
		expected int
	}{
		{"No markers", []string{"// line 1", "// line 2"}, 0, 1, 1},
		{"SPDX marker", []string{"// SPDX-License-Identifier: Apache-2.0", "// line 2"}, 0, 1, 0},
		{"Limitations marker", []string{"// line 1", "// limitations under the License."}, 0, 1, 1},
		{"Both markers", []string{"// SPDX-License-Identifier: Apache-2.0", "// limitations under the License."}, 0, 1, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := refineEndIndex(tc.lines, tc.start, tc.end); got != tc.expected {
				t.Errorf("refineEndIndex() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestIsHeaderBlock(t *testing.T) {
	testCases := []struct {
		name     string
		lines    []string
		startIdx int
		expected bool
	}{
		{"Header at top", []string{"// Copyright", "package main"}, 1, true},
		{"Empty lines before", []string{"", "// Copyright", "package main"}, 2, true},
		{"Shebang before", []string{"#!/bin/bash", "# Copyright", "echo 'hello'"}, 2, true},
		{"Go build tag before", []string{"//go:build e2e", "", "// Copyright", "package main"}, 3, true},
		{"Code before", []string{"package main", "// Copyright"}, 1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isHeaderBlock(tc.lines, tc.startIdx); got != tc.expected {
				t.Errorf("isHeaderBlock() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestProcessFile(t *testing.T) {
	tmpDir := t.TempDir()
	testCases := []struct {
		name            string
		content         string
		expectedContent string
	}{
		{
			"Go file with // comments",
			`// Copyright 2025
// Some other comment
package main`,
			`package main`,
		},
		{
			"Python file with # comments",
			`# Copyright 2025
# Some other comment
import os`,
			`import os`,
		},
		{
			"Proto file with /* */ comments",
			`/*
 * Copyright 2025
 */
syntax = "proto3";`,
			`syntax = "proto3";`,
		},
		{
			"No license header",
			`package main`,
			`package main`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use a subdirectory for each test case to avoid race conditions
			sanitizedName := strings.ReplaceAll(tc.name, " ", "_")
			sanitizedName = strings.ReplaceAll(sanitizedName, "/", "_")
			sanitizedName = strings.ReplaceAll(sanitizedName, "*", "_")
			testDir := filepath.Join(tmpDir, sanitizedName)
			if err := os.Mkdir(testDir, 0750); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}
			tmpFile := filepath.Join(testDir, "testfile")
			if err := os.WriteFile(tmpFile, []byte(tc.content), 0600); err != nil {
				t.Fatalf("Failed to write to temporary file: %v", err)
			}

			processFile(tmpFile)

			content, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read temporary file: %v", err)
			}

			if got := strings.TrimSpace(string(content)); got != strings.TrimSpace(tc.expectedContent) {
				t.Errorf("processFile() resulted in content %q, want %q", got, tc.expectedContent)
			}
		})
	}
}
