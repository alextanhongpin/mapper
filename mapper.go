package mapper

import (
	"flag"
	"fmt"
	"go/types"
	"os"
	"strings"

	"github.com/alextanhongpin/mapper/loader"
)

type Option struct {
	In       string // The input path, with the file name, e.g. yourpath/yourfile.go
	Out      string // The output path, with the mapper name, e.g. yourpath/yourfile_gen.go
	Pkg      *types.Package
	PkgName  string // The pkgName
	PkgPath  string // The pkgPath
	TypeName string // The typeName
	Suffix   string
	Type     types.Type
	DryRun   bool
}

type Generator func(opt Option) error

func New(fn Generator) error {
	typep := flag.String("type", "", "the target type name")
	suffixPtr := flag.String("suffix", "Impl", "the suffix to add to the type")
	inp := flag.String("in", os.Getenv("GOFILE"), "the input file, defaults to the file with the go:generate comment")
	outp := flag.String("out", "", "the output directory")
	dryRunp := flag.Bool("dry-run", false, "whether to print to stdout or write to file")
	pkgp := flag.String("pkg", "github.com", "the package prefix to identify the package path, override this if your packages does not reside from github.com")
	flag.Parse()

	in := loader.FullPath(*inp)

	// Allows -type=Foo,Bar
	typeNames := strings.Split(*typep, ",")
	pkg := loader.LoadPackage(loader.PackagePath(*pkgp, in)) // github.com/your-github-username/your-pkg.

	for _, typeName := range typeNames {
		out := loader.FileNameFromTypeName(*inp, *outp, typeName)
		obj := pkg.Types.Scope().Lookup(typeName)
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
			Type:     inType,
			DryRun:   *dryRunp,
		}); err != nil {
			return err
		}
	}
	return nil
}
