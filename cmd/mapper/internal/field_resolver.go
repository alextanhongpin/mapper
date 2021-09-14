package internal

import (
	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
)

var fr Resolver = new(FieldResolver)

type FieldResolver struct {
	lhs    mapper.StructField
	rhs    mapper.StructField
	assign *Assignment
}

func NewFieldResolver(name string, lhs, rhs mapper.StructField) *FieldResolver {
	fieldName0 := lhs.Name
	fieldNameN := lhs.Name
	if rhs.Tag != nil && rhs.Tag.IsAlias() {
		fieldNameN = rhs.Name
	}
	return &FieldResolver{
		lhs:    lhs,
		rhs:    rhs,
		assign: NewAssignment(name, fieldName0, fieldNameN, false),
	}
}

func (f FieldResolver) Lhs() interface{} {
	return f.lhs
}

func (f FieldResolver) Rhs() mapper.StructField {
	return f.rhs
}

func (f FieldResolver) LhsVar() *jen.Statement {
	// Output:
	// a0Name
	return f.assign.Lhs()
}

func (f FieldResolver) RhsVar() *jen.Statement {
	return f.assign.Rhs()
}

func (f FieldResolver) LhsType() *jen.Statement {
	return GenType(f.lhs.Type)
}

func (f FieldResolver) RhsType() *jen.Statement {
	return GenType(f.rhs.Type)
}

func (f *FieldResolver) Assign() {
	f.assign.Increment()
}

func (f FieldResolver) IsField() bool {
	return true
}

func (f FieldResolver) IsMethod() bool {
	return false
}

func (f FieldResolver) Tag() *mapper.Tag {
	return f.rhs.Tag
}
