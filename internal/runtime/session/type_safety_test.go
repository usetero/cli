package session

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

func TestRuntimeStructTypeSafety(t *testing.T) {
	t.Parallel()

	checkFieldNotString := func(filePath, structName, fieldName string) {
		t.Helper()
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filePath, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", filePath, err)
		}

		found := false
		ast.Inspect(node, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok || ts.Name.Name != structName {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return false
			}
			for _, f := range st.Fields.List {
				for _, name := range f.Names {
					if name.Name == fieldName {
						found = true
						if ident, ok := f.Type.(*ast.Ident); ok && ident.Name == "string" {
							t.Fatalf("%s.%s must not be string (%s)", structName, fieldName, filePath)
						}
					}
				}
			}
			return false
		})
		if !found {
			t.Fatalf("field %s.%s not found in %s", structName, fieldName, filePath)
		}
	}

	checkFieldNotString(filepath.Join("state.go"), "State", "AccountID")
	checkFieldNotString(filepath.Join("state.go"), "State", "DBPath")
	checkFieldNotString(filepath.Join("events.go"), "Event", "AccountID")
}
