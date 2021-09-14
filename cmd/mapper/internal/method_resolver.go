package internal

import (
	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
)

var mr Resolver = new(MethodResolver)

type MethodResolver struct {
	lhs    *mapper.Func
	rhs    mapper.StructField
	assign *Assignment
}

func NewMethodResolver(name string, lhs *mapper.Func, rhs mapper.StructField) *MethodResolver {
	fieldName0 := lhs.Name
	fieldNameN := lhs.Name
	if rhs.Tag != nil && rhs.Tag.IsAlias() {
		fieldNameN = rhs.Name
	}
	return &MethodResolver{
		lhs:    lhs,
		rhs:    rhs,
		assign: NewAssignment(name, fieldName0, fieldNameN, true),
	}
}

func (f MethodResolver) Lhs() interface{} {
	return f.lhs
}

func (f MethodResolver) Rhs() mapper.StructField {
	return f.rhs
}

func (f MethodResolver) LhsVar() *jen.Statement {
	// Output:
	// a0Name
	return f.assign.Lhs()
}

func (f MethodResolver) RhsVar() *jen.Statement {
	return f.assign.Rhs()
}

func (f MethodResolver) LhsType() *jen.Statement {
	return GenType(f.lhs.To.Type)
}

func (f MethodResolver) RhsType() *jen.Statement {
	return GenType(f.rhs.Type)
}

func (f *MethodResolver) Assign() {
	f.assign.Increment()
}

func (f MethodResolver) IsField() bool {
	return false
}

func (f MethodResolver) IsMethod() bool {
	return true
}

func (f MethodResolver) Tag() *mapper.Tag {
	return f.rhs.Tag
}
