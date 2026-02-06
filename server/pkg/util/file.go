// Copyright 2025 Author(s) of MCP Any
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

	// ⚡ BOLT: Optimization - Use chunk accumulation to avoid O(N^2) copying.
	// Randomized Selection from Top 5 High-Impact Targets
	return readLastNLinesOptimized(f, filesize, n)
}

func readLastNLinesOptimized(f *os.File, filesize int64, n int) ([][]byte, error) {
	const chunkSize = 1024 * 16
	buf := make([]byte, chunkSize)
	var cursor = filesize
	var chunks [][]byte
	var totalSize int
	var newlineCount int

	for cursor > 0 {
		toRead := chunkSize
		if int64(toRead) > cursor {
			toRead = int(cursor)
		}

		cursor -= int64(toRead)
		if _, err := f.Seek(cursor, io.SeekStart); err != nil {
			return nil, err
		}

		readBuf := buf[:toRead]
		if _, err := io.ReadFull(f, readBuf); err != nil {
			return nil, err
		}

		// Copy data to new slice
		chunk := make([]byte, len(readBuf))
		copy(chunk, readBuf)
		chunks = append(chunks, chunk)
		totalSize += len(chunk)

		for _, b := range chunk {
			if b == '\n' {
				newlineCount++
			}
		}

		if newlineCount >= n {
			break
		}
	}

	collected := make([]byte, totalSize)
	offset := 0
	// Chunks are stored in reverse order (End of file -> Start of file)
	// We want to reconstruct file: Start -> End.
	for i := len(chunks) - 1; i >= 0; i-- {
		copy(collected[offset:], chunks[i])
		offset += len(chunks[i])
	}

	scanner := bufio.NewScanner(bytes.NewReader(collected))
	var allLines [][]byte
	for scanner.Scan() {
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
