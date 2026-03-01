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

	graphqlRoot := filepath.Join(root, "internal", "boundary", "graphql")
	chatBoundaryRoot := filepath.Join(root, "internal", "boundary", "chat")
	powersyncBoundaryRoot := filepath.Join(root, "internal", "boundary", "powersync")
	coreRoot := filepath.Join(root, "internal", "core")
	chatRoot := filepath.Join(root, "internal", "app", "chat")

	assertNoForbiddenImports(t, graphqlRoot, []string{
		"github.com/usetero/cli/internal/app/",
	})
	assertNoForbiddenImports(t, chatBoundaryRoot, []string{
		"github.com/usetero/cli/internal/app/",
	})
	assertNoForbiddenImports(t, powersyncBoundaryRoot, []string{
		"github.com/usetero/cli/internal/app/",
	})
	assertNoForbiddenImports(t, coreRoot, []string{
		"github.com/usetero/cli/internal/app/",
		"github.com/usetero/cli/internal/boundary/graphql/",
	})
	assertNoForbiddenImportsExcept(t, chatRoot, []string{
		"github.com/usetero/cli/internal/boundary/chat",
	}, []string{
		filepath.Join("internal", "app", "chat", "messagelist", "messagelisttest"),
		filepath.Join("internal", "app", "chat", "usecase"),
	})
	assertOnlyAllowedChatClientImports(t, chatRoot, []string{
		filepath.Join("internal", "app", "chat", "messagelist", "messagelisttest"),
		filepath.Join("internal", "app", "chat", "usecase"),
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

func assertNoForbiddenImportsExcept(t *testing.T, dir string, forbiddenPrefixes []string, allowedRelPaths []string) {
	t.Helper()

	root, err := findModuleRoot(dir)
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}

	fs := token.NewFileSet()
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		f, err := parser.ParseFile(fs, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			for _, prefix := range forbiddenPrefixes {
				if !strings.HasPrefix(p, prefix) {
					continue
				}
				allowed := false
				for _, allow := range allowedRelPaths {
					if rel == allow || strings.HasPrefix(rel, allow+string(filepath.Separator)) {
						allowed = true
						break
					}
				}
				if !allowed {
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

func assertOnlyAllowedChatClientImports(t *testing.T, dir string, allowedRelPaths []string) {
	t.Helper()

	root, err := findModuleRoot(dir)
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}

	fs := token.NewFileSet()
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		f, err := parser.ParseFile(fs, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if !strings.HasPrefix(p, "github.com/usetero/cli/internal/boundary/chat") {
				continue
			}
			allowed := false
			for _, allow := range allowedRelPaths {
				if rel == allow || strings.HasPrefix(rel, allow+string(filepath.Separator)) {
					allowed = true
					break
				}
			}
			if !allowed {
				t.Errorf("%s imports boundary/chat outside allowed boundary", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
}
