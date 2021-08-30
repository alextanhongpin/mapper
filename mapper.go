package mapper

import (
	"flag"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

type Option struct {
	In       string // The input path, with the file name, e.g. yourpath/yourfile.go
	Out      string // The output path, with the mapper name, e.g. yourpath/yourfile_gen.go
	Pkg      *types.Package
	PkgName  string // The pkgName
	PkgPath  string // The pkgPath
	TypeName string // The typeName
	Suffix   string
	Type     *Type
	DryRun   bool
}

type Generator func(opt Option) error

func New(fn Generator) error {
	typePtr := flag.String("type", "", "the target type name")
	suffixPtr := flag.String("suffix", "Impl", "the suffix to add to the type")
	inPtr := flag.String("in", os.Getenv("GOFILE"), "the input file, defaults to the file with the go:generate comment")
	outPtr := flag.String("out", "", "the output directory")
	dryRunPtr := flag.Bool("dry-run", false, "whether to print to stdout or write to file")
	flag.Parse()

	in := fullPath(*inPtr)

	// Allows -type=Foo,Bar
	typeNames := strings.Split(*typePtr, ",")
	rootPkgPath := packagePath(in)

	for _, typeName := range typeNames {
		var out string
		if o := *outPtr; o == "" {
			// path/to/main.go becomes path/to/foo_gen.go
			out = filepath.Join(filepath.Dir(in), fileNameFromType(typeName))
		} else {
			if !hasExtension(o) {
				panic("mapper: out must be a valid go file")
			}
			out = fullPath(o)
		}

		pkg := LoadPackage(rootPkgPath)
		obj := LookupType(pkg, typeName)
		if obj == nil {
			panic(fmt.Sprintf("mapper: interface %s not found", typeName))
		}

		inType, ok := obj.Type().Underlying().(*types.Interface)
		if !ok {
			panic(fmt.Sprintf("mapper: %v is not an interface", obj))
		}

		if err := fn(Option{
			Pkg:      obj.Pkg(),
			PkgName:  pkg.Name,
			PkgPath:  pkg.PkgPath,
			Out:      out,
			In:       in,
			TypeName: typeName,
			Suffix:   *suffixPtr,
			Type:     NewType(inType),
			DryRun:   *dryRunPtr,
		}); err != nil {
			return err
		}
		fmt.Println("generated " + out)
	}
	return nil
}
