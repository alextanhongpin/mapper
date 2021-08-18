package mapper

import (
	"fmt"
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

func loadPackage(pkgPath string) *packages.Package {
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

func loadFunction(pkgPath, fnName string) (*packages.Package, *types.Func) {
	pkg := loadPackage(pkgPath) // github.com/your-github-username/your-pkg.
	obj := pkg.Types.Scope().Lookup(fnName)
	if obj == nil {
		panic(fmt.Sprintf("gen: func %s not found", fnName))
	}
	// Check if the type is a struct.
	funcType, ok := obj.(*types.Func)
	if !ok {
		panic(fmt.Sprintf("gen: %v is not a func", obj))
	}
	return pkg, funcType
}

func loadInterface(pkgPath, inName string) (*packages.Package, *types.Interface) {
	pkg := loadPackage(pkgPath) // github.com/your-github-username/your-pkg.
	obj := pkg.Types.Scope().Lookup(inName)
	if obj == nil {
		panic(fmt.Sprintf("gen: interface %s not found", inName))
	}

	// Check if it is a declared typed.
	if _, ok := obj.(*types.TypeName); !ok {
		log.Fatalf("gen: %v is not a named type", obj)
	}

	// Check if the type is an interface.
	inType, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		panic(fmt.Sprintf("gen: %v is not an interface", obj))
	}
	return pkg, inType
}
