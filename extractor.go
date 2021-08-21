package mapper

import (
	"fmt"
	"go/types"
)

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
		Error:   err,
		Fn:      fn,
	}
}

func ExtractInterfaceMethods(in *types.Interface) map[string]Func {
	result := make(map[string]Func)
	for i := 0; i < in.NumMethods(); i++ {
		fn := ExtractFunc(in.Method(i))
		result[fn.Name] = *fn
	}
	return result
}

func ExtractStructFields(structType *types.Struct) map[string]StructField {
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
