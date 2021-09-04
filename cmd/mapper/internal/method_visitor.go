package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type StructVisitor struct {
	fields mapper.StructFields
}

func (v *StructVisitor) Visit(T types.Type) bool {
	switch u := T.Underlying().(type) {
	case *types.Struct:
		v.fields = mapper.ExtractStructFields(u)
		return false
	}
	return true
}

func GenerateStructFields(T types.Type) mapper.StructFields {
	v := &StructVisitor{}
	_ = mapper.Walk(v, T)
	return v.fields
}
