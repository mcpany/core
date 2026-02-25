// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main checks for missing or incomplete documentation on exported symbols.
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

		checkFile(f, fset, path)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}
}

func checkFile(f *ast.File, fset *token.FileSet, path string) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				checkFuncDoc(x, fset, path)
			}
		case *ast.GenDecl:
			checkGenDecl(x, fset, path)
		}
		return true
	})
}

func checkFuncDoc(x *ast.FuncDecl, fset *token.FileSet, path string) {
	if x.Doc == nil {
		fmt.Printf("%s:%d: missing doc for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
		return
	}

	doc := x.Doc.Text()

	// Check Summary (content at start)
	if strings.TrimSpace(doc) == "" {
		fmt.Printf("%s:%d: empty doc for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
		return
	}

	// Check Parameters
	if x.Type.Params != nil && len(x.Type.Params.List) > 0 {
		// Ignore context.Context as it often doesn't need documentation or is standard
		// But strict mode might require it? AGENTS.md example didn't show context.
		// Let's assume all params need docs if strictly following "Parameters: Name, Type...".
		// Actually, let's just check for "Parameters:" label if there are params.
		if !strings.Contains(doc, "Parameters:") {
			fmt.Printf("%s:%d: missing 'Parameters:' section for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
		}
	}

	// Check Returns
	if x.Type.Results != nil && len(x.Type.Results.List) > 0 {
		if !strings.Contains(doc, "Returns:") {
			fmt.Printf("%s:%d: missing 'Returns:' section for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
		}
	}

	// Check Errors
	hasError := false
	if x.Type.Results != nil {
		for _, field := range x.Type.Results.List {
			if isErrorType(field.Type) {
				hasError = true
				break
			}
		}
	}
	if hasError {
		if !strings.Contains(doc, "Errors:") {
			fmt.Printf("%s:%d: missing 'Errors:' section for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
		}
	}

	// Check Side Effects (Always required)
	if !strings.Contains(doc, "Side Effects:") {
		fmt.Printf("%s:%d: missing 'Side Effects:' section for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
	}
}

func isErrorType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "error" {
		return true
	}
	return false
}

func checkGenDecl(x *ast.GenDecl, fset *token.FileSet, path string) {
	if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
		for _, s := range x.Specs {
			switch ts := s.(type) {
			case *ast.TypeSpec:
				if ts.Name.IsExported() {
					if x.Doc == nil && ts.Doc == nil {
						fmt.Printf("%s:%d: missing doc for type %s\n", path, fset.Position(ts.Pos()).Line, ts.Name.Name)
					}
					// Check methods in interface
					if iface, ok := ts.Type.(*ast.InterfaceType); ok {
						for _, field := range iface.Methods.List {
							if len(field.Names) > 0 && field.Names[0].IsExported() {
								if field.Doc == nil {
									fmt.Printf("%s:%d: missing doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
								} else {
                                    // Verify interface method docs too?
                                    // Interface methods are functions, so they should follow the same rules.
                                    // Construct a fake FuncDecl to reuse checkFuncDoc logic?
                                    // Or just check text.
                                    doc := field.Doc.Text()
                                    if !strings.Contains(doc, "Side Effects:") {
                                        fmt.Printf("%s:%d: missing 'Side Effects:' section for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
                                    }
                                    // Can't easily check params/returns without more work, but Side Effects is a good proxy for "Updated Docs".
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
						}
					}
				}
			}
		}
	}
}
