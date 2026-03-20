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
		fmt.Println("Usage: go run tools/check_doc_strict.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fset := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" || info.Name() == "tools" || info.Name() == "tests" || info.Name() == "test" || info.Name() == "mocks" || info.Name() == "examples" || info.Name() == "cmd" || info.Name() == "operator" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") || strings.HasSuffix(path, "zz_generated.deepcopy.go") {
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

func checkDocString(doc string, kind string, name string, path string, line int) {
    if !strings.Contains(doc, "Summary:") {
        fmt.Printf("%s:%d: missing Summary in doc for %s %s\n", path, line, kind, name)
    }

    // For functions and methods, check for Returns and Errors/Throws and Side Effects if we want to be very strict
    if kind == "function" || kind == "method" {
        if !strings.Contains(doc, "Returns:") && !strings.Contains(doc, "Returns") {
            fmt.Printf("%s:%d: missing Returns in doc for %s %s\n", path, line, kind, name)
        }
        if !strings.Contains(doc, "Errors:") && !strings.Contains(doc, "Throws") && !strings.Contains(doc, "Errors/Throws:") {
            fmt.Printf("%s:%d: missing Errors/Throws in doc for %s %s\n", path, line, kind, name)
        }
        if !strings.Contains(doc, "Side Effects:") && !strings.Contains(doc, "Side Effects") {
            fmt.Printf("%s:%d: missing Side Effects in doc for %s %s\n", path, line, kind, name)
        }
        if !strings.Contains(doc, "Parameters:") && !strings.Contains(doc, "Parameters") {
            fmt.Printf("%s:%d: missing Parameters in doc for %s %s\n", path, line, kind, name)
        }
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
                    checkDocString(x.Doc.Text(), "function", x.Name.Name, path, fset.Position(x.Pos()).Line)
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
					} else {
                        doc := ""
                        if ts.Doc != nil { doc += ts.Doc.Text() }
                        if x.Doc != nil { doc += x.Doc.Text() }
                        checkDocString(doc, "type", ts.Name.Name, path, fset.Position(ts.Pos()).Line)
                    }
					// Check methods in interface
					if iface, ok := ts.Type.(*ast.InterfaceType); ok {
						for _, field := range iface.Methods.List {
							if len(field.Names) > 0 && field.Names[0].IsExported() {
								if field.Doc == nil {
									fmt.Printf("%s:%d: missing doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
								} else {
                                    checkDocString(field.Doc.Text(), "method", field.Names[0].Name, path, fset.Position(field.Pos()).Line)
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
						} else {
                            doc := ""
                            if ts.Doc != nil { doc += ts.Doc.Text() }
                            if x.Doc != nil { doc += x.Doc.Text() }
                            checkDocString(doc, "var/const", name.Name, path, fset.Position(name.Pos()).Line)
                        }
					}
				}
			}
		}
	}
}
