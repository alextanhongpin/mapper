package mapper

import (
	"fmt"
	"go/types"
)

type FuncArg struct {
	Name string
	Type *Type
}
type FuncDto struct {
	Name    string
	PkgPath string
	From    *FuncArg
	To      *FuncArg
	Error   *Type
}

func extractNamedMethods(t *types.Named) []FuncDto {
	result := make([]FuncDto, t.NumMethods())
	for i := 0; i < t.NumMethods(); i++ {
		result[i] = *extractFunc(t.Method(i))
	}
	return result
}

func extractFunc(fn *types.Func) *FuncDto {
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
	return &FuncDto{
		Name:    fn.Name(),
		PkgPath: fn.Pkg().Path(),
		From:    from,
		To:      to,
		Error:   err,
		//Ctx:
	}
}

func extractInterfaceMethods(in *types.Interface) []FuncDto {
	result := make([]FuncDto, in.NumExplicitMethods())
	for i := 0; i < in.NumExplicitMethods(); i++ {
		result[i] = *extractFunc(in.ExplicitMethod(i))
	}
	return result
}

func extractStructFields(structType *types.Struct) map[string]StructField {
	fields := make(map[string]StructField)
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tag := structType.Tag(i)

		fields[field.Name()] = StructField{
			Name:     field.Name(),
			PkgPath:  field.Pkg().Path(),
			Exported: field.Exported(),
			Tag:      tag,
			Type:     NewType(field.Type()),
		}
	}
	return fields
}
