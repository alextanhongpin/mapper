package mapper

import (
	"fmt"
	"go/types"
)

func ExtractNamedMethods(T types.Type) map[string]Func {
	t, ok := T.(*types.Named)
	if !ok {
		panic(fmt.Sprintf("mapper: %s is not a named typed", T))
	}

	result := make(map[string]Func)
	for i := 0; i < t.NumMethods(); i++ {
		method := t.Method(i)
		if !method.Exported() {
			continue
		}
		fn := NewFunc(method)
		result[fn.Name] = *fn
	}

	return result
}

func ExtractInterfaceMethods(in *types.Interface) map[string]Func {
	result := make(map[string]Func)
	for i := 0; i < in.NumMethods(); i++ {
		fn := NewFunc(in.Method(i))
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
		if ok && tag.Ignore {
			continue
		}

		fields[key] = StructField{
			Name:     field.Name(),
			Pkg:      field.Pkg().Name(),
			PkgPath:  field.Pkg().Path(),
			Exported: field.Exported(),
			Tag:      tag,
			Type:     NewType(field.Type()),
			Var:      field,
		}
	}
	return fields
}
