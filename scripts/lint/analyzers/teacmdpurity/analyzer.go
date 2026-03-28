package teacmdpurity

import (
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/usetero/cli/scripts/lint/analyzers/analyzerutil"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "teacmdpurity",
	Doc:  "enforces that tea.Cmd closures emit messages only and do not mutate model state",
	Run:  run,
}

var repoRoot = analyzerutil.MustGetwd()

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if analyzerutil.SkipGeneratedOrTestFile(pass.Fset, file, repoRoot, nil) {
			continue
		}

		filename := filepath.ToSlash(pass.Fset.Position(file.Pos()).Filename)
		relPath := analyzerutil.RelPathFromRoot(repoRoot, filename)
		if !strings.HasPrefix(relPath, "internal/interfaces/tui/") {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.FuncLit)
			if !ok || !returnsTeaMsg(lit) {
				return true
			}

			ast.Inspect(lit.Body, func(inner ast.Node) bool {
				switch stmt := inner.(type) {
				case *ast.AssignStmt:
					if reportsModelMutation(pass, stmt.Lhs) {
						pass.Reportf(stmt.Pos(), "tea.Cmd closures must not mutate model state; return a message and update state in Update")
					}
				case *ast.IncDecStmt:
					if isModelMutationExpr(stmt.X) {
						pass.Reportf(stmt.Pos(), "tea.Cmd closures must not mutate model state; return a message and update state in Update")
					}
				}
				return true
			})

			return true
		})
	}

	return nil, nil
}

func returnsTeaMsg(lit *ast.FuncLit) bool {
	if lit == nil || lit.Type == nil || lit.Type.Results == nil || len(lit.Type.Results.List) != 1 {
		return false
	}

	result := lit.Type.Results.List[0].Type
	sel, ok := result.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "tea" && sel.Sel.Name == "Msg"
}

func reportsModelMutation(pass *analysis.Pass, lhs []ast.Expr) bool {
	for _, expr := range lhs {
		if isModelMutationExpr(expr) {
			return true
		}
	}
	return false
}

func isModelMutationExpr(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.SelectorExpr:
		return rootIdentName(v.X) == "m"
	case *ast.IndexExpr:
		return isModelMutationExpr(v.X)
	default:
		return false
	}
}

func rootIdentName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return rootIdentName(v.X)
	case *ast.IndexExpr:
		return rootIdentName(v.X)
	default:
		return ""
	}
}
