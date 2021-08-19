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
		from = &FuncArg{Name: param.Name(), Type: NewType(param.Type())}
	}
	if sig.Results().Len() > 0 {
		result := sig.Results().At(0)
		to = &FuncArg{Name: result.Name(), Type: NewType(result.Type())}
	}
	return &FuncDto{
		Name:    fn.Name(),
		PkgPath: fn.Pkg().Path(),
		From:    from,
		To:      to,
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
