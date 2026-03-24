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
        if ifStmt, ok := n.(*ast.IfStmt); ok {
            if assign, ok := ifStmt.Init.(*ast.AssignStmt); ok {
                for _, lhs := range assign.Lhs {
                    if ident, ok := lhs.(*ast.Ident); ok {
                        if ident.Name == "_" {
                            fmt.Printf("_ in if-init at %s\n", fset.Position(assign.Pos()))
                        }
                    }
                }
            }
        }
        return true
    })
}
