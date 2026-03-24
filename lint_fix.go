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
		case *ast.CallExpr:
            // Find logging.GetLogger().Error() calls without the required number of args if using slog/zap
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
                if sel.Sel.Name == "Error" {
                    if call2, ok := sel.X.(*ast.CallExpr); ok {
                        if sel2, ok := call2.Fun.(*ast.SelectorExpr); ok {
                            if sel2.Sel.Name == "GetLogger" {
                                if len(x.Args) % 2 == 0 {
                                    fmt.Printf("Error call at %s has even number of args (msg + k/v pairs): %d\n", fset.Position(n.Pos()), len(x.Args))
                                } else {
                                    fmt.Printf("Error call at %s has odd number of args (msg + k/v pairs): %d\n", fset.Position(n.Pos()), len(x.Args))
                                }
                            }
                        }
                    }
                }
			}
		}
		return true
	})
}
