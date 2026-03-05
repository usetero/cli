package chat

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

func TestStateStructTypeSafety(t *testing.T) {
	t.Parallel()

	file := filepath.Join("state.go")
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		t.Fatalf("parse state.go: %v", err)
	}

	ensureFieldNotString := func(structName, fieldName string) {
		t.Helper()
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
							t.Fatalf("%s.%s must not be string", structName, fieldName)
						}
					}
				}
			}
			return false
		})
		if !found {
			t.Fatalf("field %s.%s not found", structName, fieldName)
		}
	}

	ensureFieldNotString("State", "ConversationID")
	ensureFieldNotString("MessageView", "ID")
	ensureFieldNotString("MessageView", "Role")
}
