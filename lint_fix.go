package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
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
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Error" || sel.Sel.Name == "Info" || sel.Sel.Name == "Warn" {
					if call2, ok := sel.X.(*ast.CallExpr); ok {
						if sel2, ok := call2.Fun.(*ast.SelectorExpr); ok {
							if sel2.Sel.Name == "GetLogger" {
								if len(x.Args) > 1 && len(x.Args)%2 == 0 { // even number of args is wrong for msg + keys/values
                                    fmt.Printf("Fixing odd-number args at %s\n", fset.Position(n.Pos()))

                                    // It's usually `logging.GetLogger().Error("msg", err)` instead of `("msg", "error", err)`
                                    lastArg := x.Args[len(x.Args)-1]

                                    // Insert "error" before the last argument
                                    newArgs := make([]ast.Expr, len(x.Args)+1)
                                    copy(newArgs, x.Args[:len(x.Args)-1])
                                    newArgs[len(x.Args)-1] = &ast.BasicLit{Kind: token.STRING, Value: "\"error\""}
                                    newArgs[len(x.Args)] = lastArg

                                    x.Args = newArgs
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	os.WriteFile("server/pkg/app/api.go", buf.Bytes(), 0644)
}
