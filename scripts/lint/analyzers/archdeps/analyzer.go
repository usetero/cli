package archdeps

import (
	_ "embed"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/usetero/cli/scripts/lint/analyzers/analyzerutil"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "archdeps",
	Doc:  "enforces high-level package layer boundaries",
	Run:  run,
}

//go:embed baseline_files.txt
var baselineFilesRaw string

var repoRoot = analyzerutil.MustGetwd()
var baselineFiles = analyzerutil.ParseBaselineFiles(baselineFilesRaw)

const modulePrefix = "github.com/usetero/cli/"

type layer string

const (
	layerUnknown        layer = "unknown"
	layerCmd            layer = "cmd"
	layerInterfaces     layer = "interfaces"
	layerRuntime        layer = "runtime"
	layerDomains        layer = "domains"
	layerInfrastructure layer = "infrastructure"
)

var forbiddenImportsByLayer = map[layer]map[layer]string{
	layerUnknown:    {},
	layerCmd:        {},
	layerInterfaces: {},
	layerRuntime: {
		layerInterfaces: "runtime must not import interface packages",
	},
	layerDomains: {
		layerInterfaces:     "domains must not import interface packages",
		layerInfrastructure: "domains must not import concrete infrastructure packages",
	},
	layerInfrastructure: {
		layerInterfaces: "infrastructure must not import interface packages",
		layerRuntime:    "infrastructure must not import runtime packages",
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if analyzerutil.SkipGeneratedOrTestFile(pass.Fset, file, repoRoot, baselineFiles) {
			continue
		}

		filename := filepath.ToSlash(pass.Fset.Position(file.Pos()).Filename)
		sourceLayer := layerFromRepoPath(analyzerutil.RelPathFromRoot(repoRoot, filename))
		if sourceLayer == layerUnknown {
			continue
		}

		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			targetLayer := layerFromImportPath(importPath)
			if targetLayer == layerUnknown {
				continue
			}

			reason, forbidden := forbiddenImportsByLayer[sourceLayer][targetLayer]
			if !forbidden {
				continue
			}

			pass.Reportf(
				imp.Path.Pos(),
				"architecture dependency violation: %s (source=%s target=%s import=%q)",
				reason,
				sourceLayer,
				targetLayer,
				importPath,
			)
		}
	}

	return nil, nil
}

func layerFromImportPath(importPath string) layer {
	if !strings.HasPrefix(importPath, modulePrefix) {
		return layerUnknown
	}
	return layerFromRepoPath(strings.TrimPrefix(importPath, modulePrefix))
}

func layerFromRepoPath(path string) layer {
	switch {
	case strings.HasPrefix(path, "cmd/"):
		return layerCmd
	case strings.HasPrefix(path, "internal/interfaces/"):
		return layerInterfaces
	case strings.HasPrefix(path, "internal/runtime/"):
		return layerRuntime
	case strings.HasPrefix(path, "internal/domains/"):
		return layerDomains
	case strings.HasPrefix(path, "internal/infrastructure/"):
		return layerInfrastructure
	default:
		return layerUnknown
	}
}
