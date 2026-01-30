// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main checks for compliance with the strict documentation format.
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
		fmt.Println("Usage: go run tools/audit_doc_format.go <directory>")
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
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" || info.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip test files and generated files
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil
		}

		if checkFile(f, fset, path) {
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
	hasErrors := false
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				if err := checkFuncDoc(x); err != nil {
					fmt.Printf("%s:%d: %s\n", path, fset.Position(x.Pos()).Line, err)
					hasErrors = true
				}
			}
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				for _, s := range x.Specs {
					if ts, ok := s.(*ast.TypeSpec); ok {
						if ts.Name.IsExported() {
							// Check type documentation
							if x.Doc == nil && ts.Doc == nil {
								fmt.Printf("%s:%d: missing doc for type %s\n", path, fset.Position(ts.Pos()).Line, ts.Name.Name)
								hasErrors = true
							}

							// Check methods in interface
							if iface, ok := ts.Type.(*ast.InterfaceType); ok {
								for _, field := range iface.Methods.List {
									if len(field.Names) > 0 && field.Names[0].IsExported() {
										if field.Doc == nil {
											fmt.Printf("%s:%d: missing doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
											hasErrors = true
										} else {
											// We can't easily check params/returns for interface methods as easily as FuncDecl
											// because they are Field nodes, but we can try if we want to be strict.
											// For now, let's just ensure they are documented.
											if err := checkFieldDoc(field); err != nil {
												fmt.Printf("%s:%d: %s.%s: %s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name, err)
												hasErrors = true
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return true
	})
	return hasErrors
}

func checkFuncDoc(fn *ast.FuncDecl) error {
	if fn.Doc == nil {
		return fmt.Errorf("missing doc for function %s", fn.Name.Name)
	}

	docText := fn.Doc.Text()

	// Check for Parameters
	if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
		// Context is often excluded from doc requirements in some standards, but let's be strict if the prompt implies it.
		// "Name + Type + Constraints"
		// If the function takes parameters, "Parameters:" section is expected.
		// However, context is often skipped. Let's check if there are any non-context params.
		hasNonContextParams := false
		for _, p := range fn.Type.Params.List {
			for range p.Names { // Iterate over names (e.g. func(a, b int))
				// We'd need type checking to be sure it's context, but we can guess by name or type expression
				// Simply assuming "ctx context.Context" is skipped? The prompt says "Do not skip 'obvious' functions."
				// "Parameters: Name + Type + Constraints"
				// So I assume EVERY parameter must be documented.
				hasNonContextParams = true
			}
		}

		if hasNonContextParams {
			if !strings.Contains(docText, "Parameters:") {
				return fmt.Errorf("function %s has parameters but missing 'Parameters:' section", fn.Name.Name)
			}
		}
	}

	// Check for Returns
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		// Check if it returns only error? The prompt says "Returns: Type + Meaning (e.g., 'Returns ID on success, -1 on failure')".
		// Even for error, it should be documented.
		if !strings.Contains(docText, "Returns:") {
			return fmt.Errorf("function %s has results but missing 'Returns:' section", fn.Name.Name)
		}
	}

	return nil
}

func checkFieldDoc(field *ast.Field) error {
	if field.Doc == nil {
		return fmt.Errorf("missing doc")
	}
	docText := field.Doc.Text()

	if funcType, ok := field.Type.(*ast.FuncType); ok {
		if funcType.Params != nil && len(funcType.Params.List) > 0 {
			if !strings.Contains(docText, "Parameters:") {
				return fmt.Errorf("missing 'Parameters:' section")
			}
		}
		if funcType.Results != nil && len(funcType.Results.List) > 0 {
			if !strings.Contains(docText, "Returns:") {
				return fmt.Errorf("missing 'Returns:' section")
			}
		}
	}
	return nil
}
