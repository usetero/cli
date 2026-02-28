package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDependencyBoundaries(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root, err := findModuleRoot(wd)
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}

	apiRoot := filepath.Join(root, "internal", "api")
	coreRoot := filepath.Join(root, "internal", "core")

	assertNoForbiddenImports(t, apiRoot, []string{
		"github.com/usetero/cli/internal/app/",
	})
	assertNoForbiddenImports(t, coreRoot, []string{
		"github.com/usetero/cli/internal/app/",
		"github.com/usetero/cli/internal/api/",
	})
}

func findModuleRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func assertNoForbiddenImports(t *testing.T, dir string, forbiddenPrefixes []string) {
	t.Helper()

	fs := token.NewFileSet()
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		f, err := parser.ParseFile(fs, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			for _, prefix := range forbiddenPrefixes {
				if strings.HasPrefix(p, prefix) {
					t.Errorf("%s imports forbidden package %s", path, p)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
}
