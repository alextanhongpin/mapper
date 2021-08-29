package mapper

import (
	"fmt"
	"go/token"
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
	Error   bool
	Fn      *types.Func // Store the original
}

func NewFunc(fn *types.Func) *Func {
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		panic(fmt.Sprintf("mapper: type is not func: %v", fn))
	}

	var from, to *FuncArg
	if sig.Params().Len() > 0 {
		param := sig.Params().At(0)
		typ := NewType(param.Type())
		name := param.Name()
		if name == "" {
			name = ShortName(typ.Type)
		}
		from = NewFuncArg(name, typ, sig.Variadic())
	}

	var hasError bool
	if n := sig.Results().Len(); n > 0 {
		result := sig.Results().At(0)

		typ := NewType(result.Type())
		name := result.Name()
		if name == "" {
			name = ShortName(typ.Type)
		}
		to = NewFuncArg(name, typ, sig.Variadic())

		// Allow errors as second return value.
		if n > 1 {
			if errType := NewType(sig.Results().At(1).Type()); errType.IsError {
				hasError = true
			} else {
				// FIXME: Update error handling.
				panic("mapper: second argument must be error")
			}
		}
	}

	var pkgName, pkgPath string
	if pkg := fn.Pkg(); pkg != nil {
		pkgName = pkg.Name()
		pkgPath = pkg.Path()
	}

	return &Func{
		Name:    fn.Name(),
		Pkg:     pkgName,
		PkgPath: pkgPath,
		From:    from,
		To:      to,
		Error:   hasError,
		Fn:      fn,
	}
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

// Signature returns the conversion from A to B without error,
// pointers, slice etc.
// func mapAToB(A) B
func (f *Func) Signature() string {
	if f == nil {
		return ""
	}
	return types.TypeString(f.Fn.Type(), nil)
}

func (f *Func) Normalize() *Func {
	return NewFunc(NormFunc(f.NormalizedName(), f.Fn))
}

// RequiresInputPointer returns true if the input needs to be converted into a pointer.
func (f *Func) RequiresInputPointer(in *Type) bool {
	return !in.IsPointer && f.From.Type.IsPointer
}

// RequiresInputValue returns true if the input needs to be converted into a value.
func (f *Func) RequiresInputValue(in *Type) bool {
	return in.IsPointer && !f.From.Type.IsPointer
}

func fullName(pkgPath, name string) string {
	if pkgPath == "" {
		return name
	}
	return fmt.Sprintf("%s.%s", pkgPath, name)
}

// NormFunc generates a new func with the normalize type -
// no pointers, slice etc.
func NormFunc(name string, fn *types.Func) *types.Func {
	fullSignature := fn.Type().Underlying().(*types.Signature)

	param := NewType(fullSignature.Params().At(0).Type())
	params := types.NewTuple(types.NewVar(token.NoPos, param.ObjPkg, "", param.E))

	result := NewType(fullSignature.Results().At(0).Type())
	results := types.NewTuple(types.NewVar(token.NoPos, result.ObjPkg, "", result.E))

	sig := types.NewSignature(nil, params, results, false)
	return types.NewFunc(token.NoPos, nil, name, sig)
}

func NormFuncFromTypes(param, result *Type) *types.Func {
	params := types.NewTuple(types.NewVar(token.NoPos, param.ObjPkg, "", param.E))
	results := types.NewTuple(types.NewVar(token.NoPos, result.ObjPkg, "", result.E))

	sig := types.NewSignature(nil, params, results, false)
	return types.NewFunc(token.NoPos, nil, "", sig)
}
