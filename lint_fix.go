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
        if funcDecl, ok := n.(*ast.FuncDecl); ok {
            ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
                // look for any unhandled variables.
                // specifically check if `opts.Marshal` is fine.
                // We checked that earlier.
                return true
            })
        }
        return true
    })
}
