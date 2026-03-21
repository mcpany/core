package main

import (
	"fmt"
	"strings"
	"net/url"
)

func checkForPathTraversal(val string) error {
	if val == ".." {
		return fmt.Errorf("path traversal attempt detected")
	}
	// Check for standard traversal sequences
	if strings.HasPrefix(val, "../") || strings.HasPrefix(val, "..\\") {
		return fmt.Errorf("path traversal attempt detected")
	}
	if strings.HasSuffix(val, "/..") || strings.HasSuffix(val, "\\..") {
		return fmt.Errorf("path traversal attempt detected")
	}
	if strings.Contains(val, "/../") || strings.Contains(val, "\\..\\") || strings.Contains(val, "/..\\") || strings.Contains(val, "\\../") {
		return fmt.Errorf("path traversal attempt detected")
	}
	return nil
}

func main() {
	fmt.Println(url.PathEscape("foo/bar"))
	fmt.Println(url.PathEscape("system/services"))
}
