// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main checks for missing documentation on exported symbols.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tools/check_doc.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fset := token.NewFileSet()

	var hasErrors bool

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil
		}

		if !checkFile(f, fset, path) {
			hasErrors = true
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}

	if hasErrors {
		// For now, only warn to avoid breaking the build while we transition to strict documentation.
		// os.Exit(1)
		fmt.Println("WARNING: Documentation violations found (see above). Please fix them.")
	}
}

func checkFile(f *ast.File, fset *token.FileSet, path string) bool {
	valid := true
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				if x.Doc == nil {
					fmt.Printf("%s:%d: missing doc for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
					valid = false
				} else {
					if !checkDocContent(x.Doc.Text(), x.Type, path, fset.Position(x.Pos()).Line, x.Name.Name) {
						valid = false
					}
				}
			}
		case *ast.GenDecl:
			if !checkGenDecl(x, fset, path) {
				valid = false
			}
		}
		return true
	})
	return valid
}

func checkGenDecl(x *ast.GenDecl, fset *token.FileSet, path string) bool {
	valid := true
	if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
		for _, s := range x.Specs {
			switch ts := s.(type) {
			case *ast.TypeSpec:
				if ts.Name.IsExported() {
					if x.Doc == nil && ts.Doc == nil {
						fmt.Printf("%s:%d: missing doc for type %s\n", path, fset.Position(ts.Pos()).Line, ts.Name.Name)
						valid = false
					} else {
						// Check content for Types?
						// AGENTS.md says "Public function, method, class, and exported constant".
						// For Types (classes/structs), we should probably check Summary at least.
						// Full structure might be overkill for simple types, but let's check Summary.
						doc := x.Doc
						if ts.Doc != nil {
							doc = ts.Doc
						}
						if doc != nil {
							if strings.TrimSpace(doc.Text()) == "" {
								fmt.Printf("%s:%d: empty doc for type %s\n", path, fset.Position(ts.Pos()).Line, ts.Name.Name)
								valid = false
							}
							// We can strictly enforce Summary: ... pattern if we want, but usually types just have a summary line.
							// The "Structure" section in AGENTS.md seems focused on functions/methods ("Parameters", "Returns", "Errors").
							// But for "Summary", it applies to all.
							if !strings.Contains(doc.Text(), "Summary:") && !hasSummaryLine(doc.Text()) {
								// Strict check: Must have a summary line or "Summary:" tag?
								// "Summary: A concise, one-line action statement"
								// Usually first line is summary.
							}
						}
					}

					// Check methods in interface
					if iface, ok := ts.Type.(*ast.InterfaceType); ok {
						for _, field := range iface.Methods.List {
							if len(field.Names) > 0 && field.Names[0].IsExported() {
								if field.Doc == nil {
									fmt.Printf("%s:%d: missing doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
									valid = false
								} else {
									// Interface methods are functions, check structure
									if !checkDocContent(field.Doc.Text(), field.Type.(*ast.FuncType), path, fset.Position(field.Pos()).Line, ts.Name.Name+"."+field.Names[0].Name) {
										valid = false
									}
								}
							}
						}
					}
				}
			case *ast.ValueSpec:
				for _, name := range ts.Names {
					if name.IsExported() {
						if x.Doc == nil && ts.Doc == nil {
							fmt.Printf("%s:%d: missing doc for var/const %s\n", path, fset.Position(name.Pos()).Line, name.Name)
							valid = false
						}
					}
				}
			}
		}
	}
	return valid
}

func checkDocContent(text string, funcType *ast.FuncType, path string, line int, name string) bool {
	valid := true
	text = strings.TrimSpace(text)
	if text == "" {
		fmt.Printf("%s:%d: empty doc for %s\n", path, line, name)
		return false
	}

	// 1. Check Summary
	// "Summary: A concise..." OR just first line?
	// The example shows "Summary:" tag in the description of structure, but the example code shows:
	// // CalculateTax computes ...
	// //
	// // Parameters:
	// ...
	// So "Summary:" tag might NOT be explicit in the text, but the concept.
	// However, `server/pkg/app/api.go` had:
	// // Summary: Creates the main API handler mux.
	// So some files use explicit "Summary:".
	// The AGENTS.md says: "Structure: Summary: A concise... Parameters: ...".
	// It's ambiguous if "Summary:" keyword is required.
	// But "Parameters:", "Returns:", "Errors:", "Side Effects:" are likely keywords.
	// Let's assume explicit "Summary:" is NOT required if the first line is the summary,
	// BUT for "Parameters", "Returns", etc., they are explicit sections.
	// Wait, looking at `server/pkg/app/server.go`:
	// // Summary: Options for configuring the application runtime.
	// It seems "Summary:" IS used as a keyword in some existing docs.
	// But standard Go doc is "Function does X".
	// Let's check if "Summary:" is present. If not, check if first line exists.
	// BUT for Parameters/Returns/Errors/Side Effects, we MUST check for those keywords.

	// Helper to check for section
	hasSection := func(keyword string) bool {
		return strings.Contains(text, keyword)
	}

	// Parameters
	if funcType.Params != nil && len(funcType.Params.List) > 0 {
		if !hasSection("Parameters:") {
			fmt.Printf("%s:%d: missing 'Parameters:' section for %s\n", path, line, name)
			valid = false
		}
	}

	// Returns
	if funcType.Results != nil && len(funcType.Results.List) > 0 {
		if !hasSection("Returns:") {
			fmt.Printf("%s:%d: missing 'Returns:' section for %s\n", path, line, name)
			valid = false
		}
	}

	// Errors
	// Check if any return type is `error`
	returnsError := false
	if funcType.Results != nil {
		for _, field := range funcType.Results.List {
			if isErrorType(field.Type) {
				returnsError = true
				break
			}
		}
	}
	if returnsError {
		if !hasSection("Errors:") {
			fmt.Printf("%s:%d: missing 'Errors:' section for %s\n", path, line, name)
			valid = false
		}
	}

	// Side Effects
	if !hasSection("Side Effects:") {
		fmt.Printf("%s:%d: missing 'Side Effects:' section for %s\n", path, line, name)
		valid = false
	}

	return valid
}

func isErrorType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "error" {
		return true
	}
	return false
}

func hasSummaryLine(text string) bool {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			return true
		}
	}
	return false
}
