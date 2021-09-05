package mapper

import (
	"go/types"
)

// StructField for the example below.
//type Foo struct {
//  Name sql.NullString `json:"name"
//}
type StructField struct {
	Name string `example:"Name"`
	// Useful when the output directory doesn't match the existing ones.
	Pkg      string // e.g. yourpkg
	PkgPath  string // e.g. github.com/your-org/yourpkg
	Exported bool   // e.g. true
	Tag      *Tag   // e.g. `map:"RenameField,CustomFunction"`
	Ordinal  int    // The original position of the struct field.
	Type     types.Type
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
			// NOTE: This does not work if there are multiple
			// alias refering to the same field.
			//
			/*

				type ProductMapper interface {
					ProductToProductSummary(Products) (*ProductSummary, error)
				}

				type Products struct {
					Items []int64
				}

				// Both are referring to items.
				type ProductSummary struct {
					Items      bool  `map:",IsValidStatus"`
					TotalCount int64 `map:"Items,CountItems"`
				}
			*/
			//if tag.IsAlias() {
			//key = tag.Name
			//}
		}
		if _, exists := result[key]; exists {
			panic("duplicate key " + key)
		}
		result[key] = val
	}
	return result
}

func newStructFields(structType *types.Struct) StructFields {
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
			Type:     field.Type(),
			Ordinal:  i,
		}
	}
	return fields
}

func NewStructFields(T types.Type) StructFields {
	v := NewStructVisitor()
	_ = Walk(v, T)
	return v.fields
}

type StructVisitor struct {
	fields StructFields
}

func NewStructVisitor() *StructVisitor {
	return &StructVisitor{}
}

func (v *StructVisitor) Visit(T types.Type) bool {
	switch u := T.Underlying().(type) {
	case *types.Struct:
		v.fields = newStructFields(u)
	}
	return true
}
