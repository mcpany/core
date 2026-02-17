// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main checks for missing or non-compliant documentation on exported symbols.
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

var strictMode = os.Getenv("STRICT_DOC_CHECK") == "true"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tools/check_doc.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fset := token.NewFileSet()

	hasMissingDocs := false
	hasStructureErrors := false

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

		missing, structure := checkFile(f, fset, path)
		if missing {
			hasMissingDocs = true
		}
		if structure {
			hasStructureErrors = true
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}

	if hasMissingDocs {
		os.Exit(1)
	}
	if strictMode && hasStructureErrors {
		os.Exit(1)
	}
}

func checkFile(f *ast.File, fset *token.FileSet, path string) (bool, bool) {
	hasMissingDocs := false
	hasStructureErrors := false
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				m, s := checkFuncDoc(x, fset, path)
				if m {
					hasMissingDocs = true
				}
				if s {
					hasStructureErrors = true
				}
			}
		case *ast.GenDecl:
			m, s := checkGenDecl(x, fset, path)
			if m {
				hasMissingDocs = true
			}
			if s {
				hasStructureErrors = true
			}
		}
		return true
	})
	return hasMissingDocs, hasStructureErrors
}

func checkFuncDoc(x *ast.FuncDecl, fset *token.FileSet, path string) (bool, bool) {
	if x.Doc == nil {
		fmt.Printf("%s:%d: missing doc for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
		return true, false
	}

	doc := x.Doc.Text()
	hasStructureErrors := false

	// Check Summary (First sentence or "Summary:")
	if strings.TrimSpace(doc) == "" {
		fmt.Printf("%s:%d: empty doc for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
		return true, false
	}

	// Check Parameters
	if x.Type.Params != nil && len(x.Type.Params.List) > 0 {
		if !strings.Contains(doc, "Parameters:") {
			if strictMode {
				fmt.Printf("%s:%d: function %s missing 'Parameters:' section\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
			}
			hasStructureErrors = true
		}
	}

	// Check Returns
	if x.Type.Results != nil && len(x.Type.Results.List) > 0 {
		if !strings.Contains(doc, "Returns:") {
			if strictMode {
				fmt.Printf("%s:%d: function %s missing 'Returns:' section\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
			}
			hasStructureErrors = true
		}
	}

	return false, hasStructureErrors
}

func checkGenDecl(x *ast.GenDecl, fset *token.FileSet, path string) (bool, bool) {
	hasMissingDocs := false
	hasStructureErrors := false
	if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
		for _, s := range x.Specs {
			switch ts := s.(type) {
			case *ast.TypeSpec:
				if ts.Name.IsExported() {
					if x.Doc == nil && ts.Doc == nil {
						fmt.Printf("%s:%d: missing doc for type %s\n", path, fset.Position(ts.Pos()).Line, ts.Name.Name)
						hasMissingDocs = true
					}
					// Check methods in interface
					if iface, ok := ts.Type.(*ast.InterfaceType); ok {
						for _, field := range iface.Methods.List {
							if len(field.Names) > 0 && field.Names[0].IsExported() {
								if field.Doc == nil {
									fmt.Printf("%s:%d: missing doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
									hasMissingDocs = true
								} else {
									// Check interface method docs structure
									doc := field.Doc.Text()
									if field.Type.(*ast.FuncType).Params != nil && len(field.Type.(*ast.FuncType).Params.List) > 0 {
										if !strings.Contains(doc, "Parameters:") {
											if strictMode {
												fmt.Printf("%s:%d: interface method %s.%s missing 'Parameters:' section\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
											}
											hasStructureErrors = true
										}
									}
									if field.Type.(*ast.FuncType).Results != nil && len(field.Type.(*ast.FuncType).Results.List) > 0 {
										if !strings.Contains(doc, "Returns:") {
											if strictMode {
												fmt.Printf("%s:%d: interface method %s.%s missing 'Returns:' section\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
											}
											hasStructureErrors = true
										}
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
							hasMissingDocs = true
						}
					}
				}
			}
		}
	}
	return hasMissingDocs, hasStructureErrors
}
