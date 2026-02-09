// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

// ReadLastNLines reads the last n lines from a file.
// It is optimized to read from the end of the file.
//
// Parameters:
//   - path: string. The path to the file.
//   - n: int. The number of lines to read.
//
// Returns:
//   - [][]byte: A slice of byte slices representing the lines.
//   - error: An error if the file cannot be opened or read.
//
// Side Effects:
//   - Opens and reads the specified file from the filesystem.
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
	buf := make([]byte, chunkSize)
	var cursor = filesize

	// ⚡ BOLT: Optimized reverse file reading to avoid O(N^2) memory copying.
	// We collect chunks and join them once at the end.
	// Randomized Selection from Top 5 High-Impact Targets
	var chunks [][]byte
	newlineCount := 0

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

		// Only read the chunk we calculated
		// We re-slice buf to size needed
		readBuf := buf[:toRead]
		if _, err := io.ReadFull(f, readBuf); err != nil {
			return nil, err
		}

		// Count newlines in THIS chunk
		for _, b := range readBuf {
			if b == '\n' {
				newlineCount++
			}
		}

		// Copy chunk because buffer is reused
		chunkCopy := make([]byte, len(readBuf))
		copy(chunkCopy, readBuf)
		chunks = append(chunks, chunkCopy)

		// We need to read more than N newlines to ensure we have the start of the Nth line from the end.
		// If the file ends with a newline, we have N+1 newlines to capture N full lines.
		// If the file does not end with a newline, N newlines are sufficient.
		// To be safe, we read until we see > N newlines or hit BOF.
		if newlineCount > n {
			break
		}
	}

	// Reverse chunks to put them in correct order (since we read backwards)
	for i, j := 0, len(chunks)-1; i < j; i, j = i+1, j-1 {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	}

	// Join chunks efficiently
	collected := bytes.Join(chunks, nil)

	// Process lines manually to avoid bufio.Scanner token limit (64KB).
	// bytes.Split handles arbitrarily long lines.
	splitLines := bytes.Split(collected, []byte{'\n'})

	// If the file ends with a newline, bytes.Split produces an empty string at the end.
	// We should ignore it to match standard line reading behavior.
	if len(splitLines) > 0 && len(splitLines[len(splitLines)-1]) == 0 {
		splitLines = splitLines[:len(splitLines)-1]
	}

	// Take the last N lines
	if len(splitLines) > n {
		splitLines = splitLines[len(splitLines)-n:]
	}

	// Handle Windows CRLF: trim \r from the end of each line if present
	for i := range splitLines {
		if len(splitLines[i]) > 0 && splitLines[i][len(splitLines[i])-1] == '\r' {
			splitLines[i] = splitLines[i][:len(splitLines[i])-1]
		}
	}

	return splitLines, nil
}
