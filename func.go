package mapper

import (
	"fmt"
	"go/token"
	"go/types"
	"path"
	"strings"
	"sync"
)

type FuncArg struct {
	Name     string
	Type     types.Type
	Obj      *types.TypeName
	Variadic bool
}

func NewFuncArg(name string, T types.Type, variadic bool) *FuncArg {
	var obj *types.TypeName
	named, ok := T.(*types.Named)
	if ok {
		obj = named.Obj()
	}

	return &FuncArg{
		Name:     name,
		Type:     T,
		Obj:      obj,
		Variadic: variadic,
	}
}

type Func struct {
	Name string
	// pkg name can be different from package path, e.g. github.com/alextanhongpin/mapper/examples
	// can have package `main` instead of `examples`.
	Obj     *types.TypeName
	Pkg     string
	PkgPath string
	From    *FuncArg
	To      *FuncArg
	Error   bool
	Fn      *types.Func // Store the original

	once sync.Once
	Norm *Func
}

func NewFunc(fn *types.Func) *Func {
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		panic(fmt.Sprintf("mapper: type is not func: %v", fn))
	}

	var from, to *FuncArg
	if sig.Params().Len() > 0 {
		param := sig.Params().At(0)
		T := param.Type()
		name := param.Name()
		if name == "" {
			name = ShortName(NewUnderlyingType(T).String())
		}
		from = NewFuncArg(name, T, sig.Variadic())
	}

	var hasError bool
	if n := sig.Results().Len(); n > 0 {
		result := sig.Results().At(0)

		T := result.Type()
		name := result.Name()
		if name == "" {
			name = ShortName(NewUnderlyingType(T).String())
		}
		to = NewFuncArg(name, T, sig.Variadic())

		// Allow errors as second return value.
		if n > 1 {
			T := sig.Results().At(1).Type()
			if T.String() == "error" {
				hasError = true
			} else {
				fnstr := strings.ReplaceAll(types.TypeString(sig, (*types.Package).Name), "func", fmt.Sprintf("func %s", fn.Name()))
				panic(fmt.Errorf(`invalid function %q
hint: second return type must be error
help: replace %q with %q`,
					fnstr,
					T,
					"error",
				))
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
	_, s := path.Split(types.TypeString(NewUnderlyingType(arg.Type), (*types.Package).Name))
	s = UpperCommonInitialism(s)
	s = strings.ReplaceAll(s, ".", "")
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
	f.once.Do(func() {
		T := NormFunc(f.NormalizedName(), f.Fn)
		f.Norm = NewFunc(T)
	})
	return f.Norm
}

// RequiresInputPointer returns true if the input needs to be converted into a pointer.
func (f *Func) RequiresInputPointer(in types.Type) bool {
	return !IsPointer(in) && IsPointer(f.From.Type)
}

// RequiresInputValue returns true if the input needs to be converted into a value.
func (f *Func) RequiresInputValue(in types.Type) bool {
	return IsPointer(in) && !IsPointer(f.From.Type)
}

// NormFunc generates a new func with the normalize type -
// no pointers, slice etc.
func NormFunc(name string, fn *types.Func) *types.Func {
	fullSignature := fn.Type().Underlying().(*types.Signature)

	param := fullSignature.Params().At(0).Type()

	result := fullSignature.Results().At(0).Type()

	var paramPkg *types.Package
	namedParam, ok := param.(*types.Named)
	if ok {
		paramPkg = namedParam.Obj().Pkg()
	}

	var resultPkg *types.Package
	namedParam, ok = result.(*types.Named)
	if ok {
		resultPkg = namedParam.Obj().Pkg()
	}

	params := types.NewTuple(types.NewVar(token.NoPos, paramPkg, "", NewUnderlyingType(param)))
	results := types.NewTuple(types.NewVar(token.NoPos, resultPkg, "", NewUnderlyingType(result)))

	sig := types.NewSignature(nil, params, results, false)
	return types.NewFunc(token.NoPos, nil, name, sig)
}

func NormFuncFromTypes(param, result types.Type) *types.Func {
	var paramPkg *types.Package
	namedParam, ok := param.(*types.Named)
	if ok {
		paramPkg = namedParam.Obj().Pkg()
	}

	var resultPkg *types.Package
	namedParam, ok = result.(*types.Named)
	if ok {
		resultPkg = namedParam.Obj().Pkg()
	}

	params := types.NewTuple(types.NewVar(token.NoPos, paramPkg, "", param))
	results := types.NewTuple(types.NewVar(token.NoPos, resultPkg, "", result))

	sig := types.NewSignature(nil, params, results, false)
	return types.NewFunc(token.NoPos, nil, "", sig)
}
