package main

import (
	"fmt"
	"net/url"
)

func main() {
	baseURL, _ := url.Parse("http://example.com/api/")

    // Case: %2Ffoo
    // Path: /foo
    // RawPath: %2Ffoo

    relPath := "/foo"
    relRawPath := "%2Ffoo"

    // No prepending
    relURL := &url.URL{
        Path: relPath,
        RawPath: relRawPath,
    }

    resolvedURL := baseURL.ResolveReference(relURL)
    fmt.Printf("Base: %s\n", baseURL)
    fmt.Printf("Rel Path: %s\n", relPath)
    fmt.Printf("Rel RawPath: %s\n", relRawPath)
    fmt.Printf("Resolved: %s\n", resolvedURL.String())
    fmt.Printf("Resolved Path: %s\n", resolvedURL.Path)
    fmt.Printf("Resolved RawPath: %s\n", resolvedURL.RawPath)
}
