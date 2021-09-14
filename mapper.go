package mapper

import (
	"flag"
	"fmt"
	"go/types"
	"os"
	"sort"
	"strings"

	"github.com/alextanhongpin/mapper/loader"
)

type Option struct {
	In      string // The input path, with the file name, e.g. yourpath/yourfile.go
	Out     string // The output path, with the mapper name, e.g. yourpath/yourfile_gen.go
	Pkg     *types.Package
	PkgName string // The pkgName
	PkgPath string // The pkgPath
	Suffix  string
	DryRun  bool
	Prune   bool
	Items   []OptionItem
}

type OptionItem struct {
	Name string
	Type types.Type
	Path string
}

type TypeNames struct {
	cache map[string]bool
}

func (t TypeNames) String() string {
	return strings.Join(t.Items(), ",")
}

func (t *TypeNames) Set(val string) error {
	result := strings.Split(val, ",")
	for _, v := range result {
		t.cache[strings.TrimSpace(v)] = true
	}
	return nil
}

func (t TypeNames) Items() []string {
	var result []string
	for k := range t.cache {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

var typeNames TypeNames

func init() {
	typeNames = TypeNames{cache: make(map[string]bool)}
}

type Generator func(opt Option) error

func New(fn Generator) error {
	suffixPtr := flag.String("suffix", "Impl", "the suffix to add to the type")
	inp := flag.String("in", os.Getenv("GOFILE"), "the input file, defaults to the file with the go:generate comment")
	outp := flag.String("out", "", "the output directory")
	dryRunp := flag.Bool("dry-run", false, "whether to print to stdout or write to file")
	pkgp := flag.String("pkg", "github.com", "the package prefix to identify the package path, override this if your packages does not reside from github.com")
	prunep := flag.Bool("prune", true, "removing existing file before generating the new code")
	flag.Var(&typeNames, "type", "the target interface name")
	flag.Parse()

	in := loader.FullPath(*inp)

	// Allows -type=Foo,Bar
	pkg := loader.LoadPackage(loader.PackagePath(*pkgp, in)) // github.com/your-github-username/your-pkg.

	out := loader.FileNameFromTypeName(*inp, *outp, loader.FileName(*inp))
	opt := Option{
		//Pkg:     obj.Pkg(),
		PkgName: pkg.Name,
		PkgPath: pkg.PkgPath,
		Out:     out,
		In:      in,
		Suffix:  *suffixPtr,
		DryRun:  *dryRunp,
	}

	pruneFileIfExists := func(path string) {
		if *prunep {
			// File may not exists yet, ignore.
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				fmt.Printf("error removing file %s: %s\n", path, err)
			}
		}
	}

	for _, typeName := range typeNames.Items() {
		path := loader.FileNameFromTypeName(*inp, *outp, typeName)
		pruneFileIfExists(path)

		obj := pkg.Types.Scope().Lookup(typeName)
		if obj == nil {
			panic(fmt.Sprintf("mapper: interface %s not found", typeName))
		}

		inType, ok := obj.Type().Underlying().(*types.Interface)
		if !ok {
			panic(fmt.Sprintf("mapper: %v is not an interface", obj))
		}

		opt.Items = append(opt.Items, OptionItem{
			Path: path,
			Type: inType,
			Name: typeName,
		})
	}

	return fn(opt)
}
