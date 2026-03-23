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
	fset := token.NewFileSet()

	filepath.Walk("server", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.Contains(path, "mock_") || strings.HasSuffix(path, ".pb.go") {
			return nil
		}
		if strings.HasPrefix(path, "server/tests/") {
		    return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		for _, d := range f.Decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				if decl.Name.IsExported() {
				    if decl.Doc == nil {
				        fmt.Printf("%s:%d exported function %s should have comment or be unexported\n", path, fset.Position(decl.Pos()).Line, decl.Name.Name)
				    } else if !strings.HasPrefix(decl.Doc.Text(), decl.Name.Name+" ") && !strings.HasPrefix(decl.Doc.Text(), decl.Name.Name+"\n") && decl.Doc.Text() != decl.Name.Name {
				        fmt.Printf("%s:%d exported function %s should have comment or be unexported\n", path, fset.Position(decl.Pos()).Line, decl.Name.Name)
				    }
				}
			case *ast.GenDecl:
			    for _, spec := range decl.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
					    if s.Name.IsExported() {
					        if decl.Doc == nil {
					            fmt.Printf("%s:%d exported type %s should have comment or be unexported\n", path, fset.Position(decl.Pos()).Line, s.Name.Name)
					        } else if !strings.HasPrefix(decl.Doc.Text(), s.Name.Name+" ") && !strings.HasPrefix(decl.Doc.Text(), s.Name.Name+"\n") && decl.Doc.Text() != s.Name.Name {
					            fmt.Printf("%s:%d exported type %s should have comment or be unexported\n", path, fset.Position(decl.Pos()).Line, s.Name.Name)
					        }
					    }
					case *ast.ValueSpec:
					    for _, name := range s.Names {
					        if name.IsExported() {
					            if decl.Doc == nil {
					                fmt.Printf("%s:%d exported var/const %s should have comment or be unexported\n", path, fset.Position(decl.Pos()).Line, name.Name)
					            } else if !strings.HasPrefix(decl.Doc.Text(), name.Name+" ") && !strings.HasPrefix(decl.Doc.Text(), name.Name+"\n") && decl.Doc.Text() != name.Name {
					                fmt.Printf("%s:%d exported var/const %s should have comment or be unexported\n", path, fset.Position(decl.Pos()).Line, name.Name)
					            }
					        }
					    }
					}
				}
			}
		}

		return nil
	})
}
