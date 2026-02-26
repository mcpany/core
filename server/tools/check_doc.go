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
	hasErrors := false

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
		os.Exit(1)
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
					if !checkDocStructure(x.Doc.Text(), x.Type, path, fset.Position(x.Pos()).Line, x.Name.Name) {
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

func checkDocStructure(doc string, funcType *ast.FuncType, path string, line int, name string) bool {
	valid := true
	// Basic checks
	if !strings.Contains(doc, "Summary:") && !strings.Contains(doc, name) {
		// Assume first line is summary if not explicit, but strict mode prefers explicit or clear structure.
	}

	// Parameters check
	if funcType.Params != nil && len(funcType.Params.List) > 0 {
		if !strings.Contains(doc, "Parameters:") {
			fmt.Printf("%s:%d: function %s missing 'Parameters:' section\n", path, line, name)
			valid = false
		}
	}

	// Returns check
	if funcType.Results != nil && len(funcType.Results.List) > 0 {
		if !strings.Contains(doc, "Returns:") {
			fmt.Printf("%s:%d: function %s missing 'Returns:' section\n", path, line, name)
			valid = false
		}
	}

	// Errors check
	hasError := false
	if funcType.Results != nil {
		for _, field := range funcType.Results.List {
			if isErrorType(field.Type) {
				hasError = true
				break
			}
		}
	}
	if hasError && !strings.Contains(doc, "Errors:") {
		fmt.Printf("%s:%d: function %s returns error but missing 'Errors:' section\n", path, line, name)
		valid = false
	}

	// Side Effects check
	if !strings.Contains(doc, "Side Effects:") {
		fmt.Printf("%s:%d: function %s missing 'Side Effects:' section\n", path, line, name)
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
					}
					// Check methods in interface
					if iface, ok := ts.Type.(*ast.InterfaceType); ok {
						for _, field := range iface.Methods.List {
							if len(field.Names) > 0 && field.Names[0].IsExported() {
								if field.Doc == nil {
									fmt.Printf("%s:%d: missing doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
									valid = false
								} else {
									if funcType, ok := field.Type.(*ast.FuncType); ok {
										if !checkDocStructure(field.Doc.Text(), funcType, path, fset.Position(field.Pos()).Line, ts.Name.Name+"."+field.Names[0].Name) {
											valid = false
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
							valid = false
						}
					}
				}
			}
		}
	}
	return valid
}
