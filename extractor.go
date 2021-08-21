package mapper

import (
	"fmt"
	"go/types"
	"path"
	"strings"
)

type FuncArg struct {
	Name string
	Type *Type
}

type Func struct {
	Name string
	// pkg name can be different from package path, e.g. github.com/alextanhongpin/mapper/examples
	// can have package `main` instead of `examples`.
	Pkg     string
	PkgPath string
	From    *FuncArg
	To      *FuncArg
	Error   *Type
	Fn      *types.Func // Store the original
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

func (f *Func) NormalizedSignature() string {
	returnTuple := fullName(f.To.Type.PkgPath, f.To.Type.Type)
	if f.Error != nil {
		returnTuple = fmt.Sprintf("(%s, error)", returnTuple)
	}

	return fmt.Sprintf("func %s(%s) %s",
		f.NormalizedName(),
		fullName(f.From.Type.PkgPath, f.From.Type.Type),
		returnTuple,
	)
}

func fullName(pkgPath, name string) string {
	if pkgPath == "" {
		return name
	}
	return fmt.Sprintf("%s.%s", pkgPath, name)
}

func extractNamedMethods(t *types.Named) map[string]Func {
	result := make(map[string]Func)
	for i := 0; i < t.NumMethods(); i++ {
		fn := ExtractFunc(t.Method(i))
		result[fn.Name] = *fn
	}
	return result
}

func ExtractFunc(fn *types.Func) *Func {
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
		from = &FuncArg{Name: name, Type: typ}
	}

	var err *Type
	if n := sig.Results().Len(); n > 0 {
		result := sig.Results().At(0)

		typ := NewType(result.Type())
		name := result.Name()
		if name == "" {
			name = ShortName(typ.Type)
		}
		to = &FuncArg{Name: name, Type: typ}

		// Allow errors as second return value.
		if n > 1 {
			if errType := NewType(sig.Results().At(1).Type()); errType.IsError {
				err = errType
			}
		}
	}
	return &Func{
		Name:    fn.Name(),
		Pkg:     fn.Pkg().Name(),
		PkgPath: fn.Pkg().Path(),
		From:    from,
		To:      to,
		Error:   err,
		Fn:      fn,
	}
}

func extractInterfaceMethods(in *types.Interface) map[string]Func {
	result := make(map[string]Func)
	for i := 0; i < in.NumMethods(); i++ {
		fn := ExtractFunc(in.Method(i))
		result[fn.Name] = *fn
	}
	return result
}

func extractStructFields(structType *types.Struct) map[string]StructField {
	fields := make(map[string]StructField)
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		key := field.Name()

		tag, ok := NewTag(structType.Tag(i))
		if ok && tag.IsAlias() {
			key = tag.Name
		}

		fields[key] = StructField{
			Name:     field.Name(),
			PkgPath:  field.Pkg().Path(),
			Exported: field.Exported(),
			Tag:      tag,
			Type:     NewType(field.Type()),
			Var:      field,
		}
	}
	return fields
}
