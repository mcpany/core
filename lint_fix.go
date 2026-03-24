package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "server/pkg/app/api.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

    ast.Inspect(f, func(n ast.Node) bool {
        if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Name.Name == "handleTools" {
            ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
                if assign, ok := n.(*ast.AssignStmt); ok {
                    for _, lhs := range assign.Lhs {
                        if ident, ok := lhs.(*ast.Ident); ok {
                            if ident.Name == "_" { // look for ANY _ assignment
                                if len(assign.Rhs) > 0 {
                                    fmt.Printf("_ assign at %s\n", fset.Position(assign.Pos()))
                                }
                            }
                        }
                    }
                }
                return true
            })
        }
        return true
    })
}
