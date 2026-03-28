package analyzerutil

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func ParseBaselineFiles(raw string) map[string]bool {
	out := make(map[string]bool)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[line] = true
	}
	return out
}

func MustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.ToSlash(wd)
}

func RelPathFromRoot(root, path string) string {
	if root == "" {
		return filepath.ToSlash(path)
	}
	path = filepath.ToSlash(path)
	root = strings.TrimSuffix(filepath.ToSlash(root), "/")
	if strings.HasPrefix(path, root+"/") {
		return strings.TrimPrefix(path, root+"/")
	}
	return path
}

func SkipGeneratedOrTestFile(fset *token.FileSet, file *ast.File, repoRoot string, baselineFiles map[string]bool) bool {
	filename := filepath.ToSlash(fset.Position(file.Pos()).Filename)
	relPath := RelPathFromRoot(repoRoot, filename)

	if strings.HasSuffix(filename, "_test.go") {
		return true
	}
	if strings.HasSuffix(filename, ".generated.go") || strings.HasSuffix(filename, ".pb.go") || strings.HasSuffix(filename, ".pb.gw.go") {
		return true
	}
	if baselineFiles[relPath] {
		return true
	}
	if IsSupportTestPath(filename) {
		return true
	}
	if HasGeneratedHeader(file) {
		return true
	}

	return false
}

func IsSupportTestPath(filename string) bool {
	for _, part := range strings.Split(filepath.ToSlash(filename), "/") {
		if strings.HasSuffix(part, "test") || strings.HasSuffix(part, "tests") {
			return true
		}
	}
	return false
}

func HasGeneratedHeader(file *ast.File) bool {
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			if strings.Contains(text, "Code generated") && strings.Contains(text, "DO NOT EDIT") {
				return true
			}
		}
	}
	return false
}
