package mapper

import (
	"go/types"
)

func extractStructFields(structType *types.Struct) map[string]StructField {
	fields := make(map[string]StructField)
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tag := structType.Tag(i)

		fields[field.Name()] = StructField{
			Name:     field.Name(),
			PkgPath:  field.Pkg().Path(),
			Exported: field.Exported(),
			Field:    NewField(field.Type()),
			Tag:      tag,
		}
	}
	return fields
}

func compareStructFields(src map[string]StructField, tgt map[string]StructField) bool {
	if len(src) != len(tgt) {
		return false
	}
	for key := range src {
		if _, exist := tgt[key]; !exist {
			return false
		}
	}
	return true
}
