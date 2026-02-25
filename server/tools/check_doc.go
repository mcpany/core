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
				if x.Doc == nil {
					fmt.Printf("%s:%d: missing doc for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
				} else {
					checkDocStructure(x.Doc.Text(), x.Type, path, fset.Position(x.Pos()).Line, x.Name.Name)
				}
			}
		case *ast.GenDecl:
			checkGenDecl(x, fset, path)
		}
		return true
	})
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
									checkDocStructure(field.Doc.Text(), field.Type.(*ast.FuncType), path, fset.Position(field.Pos()).Line, ts.Name.Name+"."+field.Names[0].Name)
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

func checkDocStructure(doc string, funcType *ast.FuncType, path string, line int, name string) {
	// 1. Summary
	lines := strings.Split(strings.TrimSpace(doc), "\n")
	if len(lines) == 0 {
		fmt.Printf("%s:%d: empty doc for function %s\n", path, line, name)
		return
	}
	// We assume first line is Summary.

	// 2. Parameters
	hasParams := false
	if funcType.Params != nil && len(funcType.Params.List) > 0 {
		hasParams = true
	}
	if hasParams && !strings.Contains(doc, "Parameters:") {
		fmt.Printf("%s:%d: missing 'Parameters:' section in doc for function %s\n", path, line, name)
	}

	// 3. Returns
	hasReturns := false
	hasError := false
	if funcType.Results != nil && len(funcType.Results.List) > 0 {
		hasReturns = true
		for _, field := range funcType.Results.List {
			// Basic check for error return
			if isErrorType(field.Type) {
				hasError = true
			}
		}
	}
	if hasReturns && !strings.Contains(doc, "Returns:") {
		fmt.Printf("%s:%d: missing 'Returns:' section in doc for function %s\n", path, line, name)
	}

	// 4. Errors
	if hasError && !strings.Contains(doc, "Errors:") {
		fmt.Printf("%s:%d: missing 'Errors:' section in doc for function %s (returns error)\n", path, line, name)
	}

	// 5. Side Effects (Mandatory per spec)
	if !strings.Contains(doc, "Side Effects:") {
		fmt.Printf("%s:%d: missing 'Side Effects:' section in doc for function %s\n", path, line, name)
	}
}

func isErrorType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "error" {
		return true
	}
	// Could check for qualified name like pkg.Error but error is usually just error
	return false
}
