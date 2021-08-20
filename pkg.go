package mapper

import (
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func stripExtension(path string) string {
	if ext := filepath.Ext(path); ext != "" {
		base := filepath.Base(path)
		path = path[:len(path)-len(base)]
	}
	return path
}

func stripSlash(path string) string {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}

// packagePath returns the github package path from any given path,
// e.g. path/to/github.com/your-repo/your-pkg returns github.com/your-repo/your-pkg
// If your package is not hosted on github, you may need to override $PKG to
// set the prefix of your package.
func packagePath(path string) string {
	path = stripExtension(path)
	path = stripSlash(path)
	pkg := os.Getenv("PKG")
	if pkg == "" {
		pkg = "github.com"
	}
	idx := strings.Index(path, pkg)
	return path[idx:]
}

// packageName returns the base package name from a given package path.
// github.com/alextanhongpin/mypackage -> mypackage
func packageName(pkgPath string) string {
	return filepath.Base(packagePath(pkgPath))
}

func LoadPackage(pkgPath string) *packages.Package {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		log.Fatalf("failed to load package: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}
	return pkgs[0]
}

func LookupType(pkg *packages.Package, typeName string) types.Object {
	return pkg.Types.Scope().Lookup(typeName)
}
