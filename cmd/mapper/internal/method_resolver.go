package internal

import (
	"github.com/alextanhongpin/mapper"
	"github.com/dave/jennifer/jen"
)

var mr Resolver = new(MethodResolver)

type MethodResolver struct {
	name  string
	lhs   *mapper.Func
	rhs   mapper.StructField
	count int
}

func NewMethodResolver(name string, lhs *mapper.Func, rhs mapper.StructField) *MethodResolver {
	return &MethodResolver{
		name: name,
		lhs:  lhs,
		rhs:  rhs,
	}
}

func (f MethodResolver) Lhs() interface{} {
	return f.lhs
}

func (f MethodResolver) fieldName() string {
	if f.rhs.Tag != nil && f.rhs.Tag.IsAlias() {
		return f.rhs.Name
	}
	return f.lhs.Name
}

func (f MethodResolver) Rhs() mapper.StructField {
	return f.rhs
}

func (f MethodResolver) LhsVar() *jen.Statement {
	// Output:
	// a0Name
	return jen.Id(argsWithIndex(f.name, f.count) + f.fieldName())
}

func (f MethodResolver) RhsVar() *jen.Statement {
	if f.count == 0 {
		// Output:
		// a0.Name()
		return jen.Id(argsWithIndex(f.name, f.count)).Dot(f.lhs.Name).Call()
	}
	// Output:
	// a0Name
	return jen.Id(argsWithIndex(f.name, f.count-1) + f.fieldName())
}

func (f MethodResolver) LhsType() *jen.Statement {
	return GenType(f.lhs.To.Type)
}

func (f MethodResolver) RhsType() *jen.Statement {
	return GenType(f.rhs.Type)
}

func (f *MethodResolver) Assign() {
	f.count++
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
