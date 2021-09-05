package internal

import (
	"go/types"

	"github.com/alextanhongpin/mapper"
)

type FuncParamVisitor struct {
	fields       mapper.StructFields
	methods      map[string]*mapper.Func
	isCollection bool
}

func NewFuncParamVisitor() *FuncParamVisitor {
	return &FuncParamVisitor{}
}

func (v *FuncParamVisitor) Visit(T types.Type) bool {
	switch u := T.(type) {
	case *types.Array, *types.Slice:
		v.isCollection = true
	case *types.Named:
		v.methods = mapper.ExtractNamedMethods(u)
	case *types.Struct:
		v.fields = mapper.ExtractStructFields(u).WithTags()
		return false
	default:
		panic("not handled")
	}
	return true
}

func (v FuncParamVisitor) FieldByName(name string) (mapper.StructField, bool) {
	field, ok := v.fields[name]
	return field, ok
}

func (v FuncParamVisitor) MethodByName(name string) (*mapper.Func, bool) {
	method, ok := v.methods[name]
	return method, ok
}

// HasError returns true if the LHS field is method call
// and returns error as the result tuple.
func (v FuncParamVisitor) HasError() bool {
	for _, met := range v.methods {
		if met.Error {
			return true
		}
	}
	return false
}
