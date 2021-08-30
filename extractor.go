package mapper

import (
	"fmt"
	"go/types"
)

func ExtractNamedMethods(T types.Type) map[string]*Func {
	t, ok := T.(*types.Named)
	if !ok {
		panic(fmt.Sprintf("mapper: %s is not a named typed", T))
	}

	result := make(map[string]*Func)
	for i := 0; i < t.NumMethods(); i++ {
		method := t.Method(i)
		if !method.Exported() {
			continue
		}
		fn := NewFunc(method)
		result[fn.Name] = fn
	}

	return result
}

func ExtractInterfaceMethods(in *types.Interface) map[string]*Func {
	result := make(map[string]*Func)
	for i := 0; i < in.NumMethods(); i++ {
		fn := NewFunc(in.Method(i))
		result[fn.Name] = fn
	}
	return result
}

type StructFields map[string]StructField

func (s StructFields) WithTags() StructFields {
	result := make(StructFields)
	for key, val := range s {
		tag := val.Tag
		if tag != nil {
			if tag.Ignore {
				continue
			}
			if tag.IsAlias() {
				key = tag.Name
			}
		}
		result[key] = val
	}
	return result
}

func ExtractStructFields(structType *types.Struct) StructFields {
	fields := make(StructFields)
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		key := field.Name()

		tag, _ := NewTag(structType.Tag(i))

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
