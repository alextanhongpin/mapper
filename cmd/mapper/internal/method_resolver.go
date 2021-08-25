package internal

import (
	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
)

var mr Resolver = new(MethodResolver)

type MethodResolver struct {
	name string
	// If the custom `map` tag is not defined, this will be nil.
	field *mapper.StructField
	lhs   mapper.Func
	rhs   mapper.StructField
	count int
}

func NewMethodResolver(name string, field *mapper.StructField, lhs mapper.Func, rhs mapper.StructField) *MethodResolver {
	return &MethodResolver{
		name: name,
		lhs:  lhs,
		rhs:  rhs,
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
	return jen.Id(argsWithIndex(f.name, f.count) + f.lhs.Name).Clone()
}

func (f MethodResolver) RhsVar() *jen.Statement {
	if f.count == 0 {
		// Output:
		// a0.Name()
		return jen.Id(argsWithIndex(f.name, f.count)).Dot(f.lhs.Name).Call().Clone()
	}
	// Output:
	// a0Name
	return jen.Id(argsWithIndex(f.name, f.count-1)).Clone()
}

func (f MethodResolver) LhsType() *jen.Statement {
	return GenType(f.lhs.To.Type)
}

func (f MethodResolver) RhsType() *jen.Statement {
	return GenType(f.rhs.Type)
}

func (f MethodResolver) Assign() {
	f.count++
}

func (f MethodResolver) IsField() bool {
	return false
}

func (f MethodResolver) IsMethod() bool {
	return true
}

func (f MethodResolver) Tag() *mapper.Tag {
	if f.field == nil {
		return nil
	}
	return f.field.Tag
}
