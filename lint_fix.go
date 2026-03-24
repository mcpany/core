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
            for _, lhs := range assign.Lhs {
                if ident, ok := lhs.(*ast.Ident); ok {
                    if ident.Name == "_" && len(assign.Lhs) > 1 {
                        // Check if the rhs is a function call that might return an error
                        if len(assign.Rhs) > 0 {
                            if call, ok := assign.Rhs[0].(*ast.CallExpr); ok {
                                // Just a simple heuristic to spot ignored errors
                                fmt.Printf("Potentially ignored error at %s\n", fset.Position(assign.Pos()))
                                _ = call
                            }
                        }
                    }
                }
            }
        }
        return true
    })
}
