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
		if stmt, ok := n.(*ast.AssignStmt); ok {
			for _, rhs := range stmt.Rhs {
				if ident, ok := rhs.(*ast.Ident); ok {
					if ident.Name == "req" {
						fmt.Printf("found req rhs at %s\n", fset.Position(ident.Pos()))
					}
				}
			}
		}

		if call, ok := n.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok {
				if ident.Name == "append" {
					if len(call.Args) > 0 {
						if argIdent, ok := call.Args[0].(*ast.Ident); ok {
							if argIdent.Name == "buf" {
								//fmt.Printf("append to buf at %s\n", fset.Position(call.Pos()))
							}
						}
					}
				}
			}
		}
		return true
	})
}
