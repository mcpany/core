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
		switch x := n.(type) {
		case *ast.AssignStmt:
            if len(x.Lhs) > 0 {
                if ident, ok := x.Lhs[0].(*ast.Ident); ok {
                    fmt.Printf("Assignment to %s at %s\n", ident.Name, fset.Position(n.Pos()))
                }
            }
		}
		return true
	})
}
