package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// ExitCheckAnalyzer analyzer for checking os.Exit from main functions main package.
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for exits from main functions main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	expr := func(x *ast.ExprStmt) {
		call, ok := x.X.(*ast.CallExpr)
		if !ok {
			return
		}
		if sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(id.NamePos, "os.Exit from main function main package")
			}
		}
	}
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				if x.Name.Name != "main" {
					return false
				}
			case *ast.ExprStmt:
				expr(x)
			}
			return true
		})
	}
	return nil, nil
}
