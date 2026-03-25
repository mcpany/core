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
        if assign, ok := n.(*ast.AssignStmt); ok {
            if len(assign.Lhs) > 1 {
                for _, lhs := range assign.Lhs {
                    if ident, ok := lhs.(*ast.Ident); ok {
                        if ident.Name == "err" { // look for explicit err declaration that isn't handled

                        }
                    }
                }
            }
        }
        return true
    })
}
