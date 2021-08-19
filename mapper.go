package mapper

import (
	"flag"
	"fmt"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Option struct {
	In       string // The input path, with the file name, e.g. yourpath/yourfile.go
	Out      string // The output path, with the gen name, e.g. yourpath/yourfile_gen.go
	PkgName  string // The pkgName
	PkgPath  string // The pkgPath
	TypeName string // The typeName
	Type     *Type
}

type Generator func(opt Option) error

func New(fn Generator) error {
	typePtr := flag.String("type", "", "the target type name")
	inPtr := flag.String("in", os.Getenv("GOFILE"), "the input file, defaults to the file with the go:generate comment")
	outPtr := flag.String("out", "", "the output directory")
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

		pkg, inType := loadInterface(rootPkgPath, typeName) // github.com/your-github-username/your-pkg.
		log.Printf("inPkg: %v\n", pkg)
		log.Printf("inType: %v\n", NewType(inType))

		//fnPkg, fnType := loadFunction(rootPkgPath, "CustomConverter")
		//log.Printf("fnPkg: %v\n", fnPkg)
		//log.Printf("fnType: %v\n", fnType)

		if err := fn(Option{
			PkgName:  pkg.Name,
			PkgPath:  pkg.PkgPath,
			Out:      out,
			In:       in,
			TypeName: typeName,
			Type:     NewType(inType),
		}); err != nil {
			return err
		}
	}
	return nil
}

func debugVar(name string, v *types.Var) {
	fmt.Printf("%s: %v\n", name, v)
	fmt.Printf("%s.Anonymous(): %v\n", name, v.Anonymous())
	fmt.Printf("%s.IsField(): %v\n", name, v.IsField())
	fmt.Printf("%s.Exported(): %v\n", name, v.Exported())
	fmt.Printf("%s.Name(): %v\n", name, v.Name())
	fmt.Printf("%s.Pkg(): %v\n", name, v.Pkg())
	fmt.Printf("%s.Type(): %v\n", name, v.Type())
	fmt.Printf("%s.String(): %v\n", name, v.String())
	fmt.Printf("%s.NewField(): %+v\n", name, NewType(v.Type()))

	switch t := v.Type().Underlying().(type) {
	case *types.Pointer:
		fmt.Println("is pointer", t)
	case *types.Struct:
		fmt.Println("is struct", t)
	case *types.Slice:
		fmt.Println("is slice", t)
	case *types.Array:
		fmt.Println("is array", t)
	case *types.Map:
		fmt.Println("is map", t)
	default:
		fmt.Println("is unknown", t)
	}
}
