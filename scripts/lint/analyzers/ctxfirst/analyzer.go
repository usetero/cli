package ctxfirst

import (
	_ "embed"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/usetero/cli/scripts/lint/analyzers/analyzerutil"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "ctxfirst",
	Doc:  "enforces context.Context as the first function or method parameter when present",
	Run:  run,
}

//go:embed baseline_files.txt
var baselineFilesRaw string

var repoRoot = analyzerutil.MustGetwd()
var baselineFiles = analyzerutil.ParseBaselineFiles(baselineFilesRaw)

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if analyzerutil.SkipGeneratedOrTestFile(pass.Fset, file, repoRoot, baselineFiles) {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				checkSignature(pass, fn.Type, fn.Pos())
			case *ast.FuncLit:
				checkSignature(pass, fn.Type, fn.Pos())
			}
			return true
		})
	}

	return nil, nil
}

func checkSignature(pass *analysis.Pass, fnType *ast.FuncType, pos token.Pos) {
	if fnType == nil || fnType.Params == nil || len(fnType.Params.List) == 0 {
		return
	}

	first := fnType.Params.List[0]
	firstIsContext := isContextType(pass.TypesInfo.TypeOf(first.Type))

	for _, param := range fnType.Params.List {
		if !isContextType(pass.TypesInfo.TypeOf(param.Type)) {
			continue
		}
		if firstIsContext {
			return
		}
		pass.Reportf(pos, "context.Context must be the first parameter when present")
		return
	}
}

func isContextType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return obj.Pkg().Path() == "context" && obj.Name() == "Context"
}
