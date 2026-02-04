// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

// ReadLastNLines reads the last n lines from a file.
// It is optimized to read from the end of the file.
func ReadLastNLines(path string, n int) ([][]byte, error) {
	if n <= 0 {
		return nil, nil
	}

	f, err := os.Open(path) //nolint:gosec // Intended file inclusion
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	filesize := stat.Size()
	if filesize == 0 {
		return nil, nil
	}

	var lines [][]byte

	// If file is small, just read it all
	if filesize < 64*1024 { // 64KB arbitrarily small enough to read fully
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Bytes())
		}
		if len(lines) > n {
			return lines[len(lines)-n:], nil
		}
		return lines, nil
	}

	// Seek backwards
	// Use a reasonable chunk size
	const chunkSize = 1024 * 16
	var cursor = filesize

	// ⚡ BOLT: Optimized to avoid quadratic memory allocation when prepending chunks.
	// Randomized Selection from Top 5 High-Impact Targets
	var chunks [][]byte
	totalNewlines := 0
	totalBytes := 0

	for cursor > 0 {
		toRead := chunkSize
		if int64(toRead) > cursor {
			toRead = int(cursor)
		}

		cursor -= int64(toRead)
		_, err = f.Seek(cursor, io.SeekStart)
		if err != nil {
			return nil, err
		}

		// Read chunk
		chunk := make([]byte, toRead)
		if _, err := io.ReadFull(f, chunk); err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
		totalBytes += len(chunk)

		// Count newlines in this chunk
		for _, b := range chunk {
			if b == '\n' {
				totalNewlines++
			}
		}

		if totalNewlines >= n {
			break
		}
	}

	// Assemble collected buffer from chunks in reverse order
	collected := make([]byte, totalBytes)
	offset := 0
	for i := len(chunks) - 1; i >= 0; i-- {
		copy(collected[offset:], chunks[i])
		offset += len(chunks[i])
	}

	// Now process 'collected'
	scanner := bufio.NewScanner(bytes.NewReader(collected))
	var allLines [][]byte
	for scanner.Scan() {
		// Copy bytes because scanner reuses buffer
		b := scanner.Bytes()
		tmp := make([]byte, len(b))
		copy(tmp, b)
		allLines = append(allLines, tmp)
	}

	if len(allLines) > n {
		return allLines[len(allLines)-n:], nil
	}
	return allLines, nil
}
