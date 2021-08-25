package internal

import (
	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
)

var fr Resolver = new(FieldResolver)

type FieldResolver struct {
	name  string
	lhs   mapper.StructField
	rhs   mapper.StructField
	count int
}

func NewFieldResolver(name string, lhs, rhs mapper.StructField) *FieldResolver {
	return &FieldResolver{
		name: name,
		rhs:  rhs,
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
	return jen.Id(argsWithIndex(f.name, f.count) + f.lhs.Name).Clone()
}

func (f FieldResolver) RhsVar() *jen.Statement {
	if f.count == 0 {
		// Output:
		// a0.Name
		return jen.Id(argsWithIndex(f.name, f.count)).Dot(f.lhs.Name).Clone()
	}
	// Output:
	// a0Name
	return jen.Id(argsWithIndex(f.name, f.count-1)).Clone()
}

func (f FieldResolver) LhsType() *jen.Statement {
	return GenType(f.lhs.Type)
}

func (f FieldResolver) RhsType() *jen.Statement {
	return GenType(f.rhs.Type)
}

func (f FieldResolver) Assign() {
	f.count++
}

func (f FieldResolver) IsField() bool {
	return true
}

func (f FieldResolver) IsMethod() bool {
	return false
}

func (f FieldResolver) Tag() *mapper.Tag {
	return f.lhs.Tag
}
