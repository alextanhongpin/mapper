package mapper

import (
	"fmt"
	"go/types"
	"path"
	"strings"
)

type FuncArg struct {
	Name     string
	Type     *Type
	Variadic bool
}

func NewFuncArg(name string, T *Type, variadic bool) *FuncArg {
	return &FuncArg{
		Name:     name,
		Type:     T,
		Variadic: variadic,
	}
}

type Func struct {
	Name string
	// pkg name can be different from package path, e.g. github.com/alextanhongpin/mapper/examples
	// can have package `main` instead of `examples`.
	Pkg     string
	PkgPath string
	From    *FuncArg
	To      *FuncArg
	// TODO: Change this to boolean.
	Error *Type
	Fn    *types.Func // Store the original
}

func (f *Func) normalizedArg(arg *FuncArg) string {
	_, s := path.Split(fullName(arg.Type.Pkg, arg.Type.Type))
	s = strings.ReplaceAll(s, ".", "")
	s = UpperCommonInitialism(s)
	return s
}

func (f *Func) NormalizedName() string {
	in := f.normalizedArg(f.From)
	out := f.normalizedArg(f.To)
	return fmt.Sprintf("map%sTo%s", in, out)
}

// NormalizedSignature returns the conversion from A to B without error,
// pointers, slice etc.
// func mapAToB(A) B
func (f *Func) NormalizedSignature() string {
	return fmt.Sprintf("func %s(%s) %s",
		f.NormalizedName(),
		fullName(f.From.Type.PkgPath, f.From.Type.Type),
		fullName(f.To.Type.PkgPath, f.To.Type.Type),
	)
}

func (f *Func) PrettySignature() string {
	return fmt.Sprintf("func %s(%s) %s",
		f.Name,
		fullName(f.From.Type.Pkg, f.From.Type.Type),
		fullName(f.To.Type.Pkg, f.To.Type.Type),
	)
}

func fullName(pkgPath, name string) string {
	if pkgPath == "" {
		return name
	}
	return fmt.Sprintf("%s.%s", pkgPath, name)
}
