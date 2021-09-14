package internal

import "github.com/dave/jennifer/jen"

type Assignment struct {
	count       int
	fieldPrefix string
	fieldName0  string
	fieldNameN  string
	method      bool
}

func NewAssignment(fieldPrefix, fieldName0, fieldNameN string, method bool) *Assignment {
	return &Assignment{
		fieldPrefix: fieldPrefix,
		fieldName0:  fieldName0,
		fieldNameN:  fieldNameN,
		method:      method,
	}
}

func (a *Assignment) Increment() {
	a.count++
}

func (a *Assignment) Lhs() *jen.Statement {
	return jen.Id(argsWithIndex(a.fieldPrefix, a.count) + a.fieldNameN)
}

/*
	NOTE: For scenario with similar alias.

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
func (a *Assignment) Rhs() *jen.Statement {
	var fieldName string
	switch a.count {
	case 0:
		fieldName = a.fieldName0
		return jen.Id(argsWithIndex(a.fieldPrefix, a.count)).Dot(fieldName).Do(func(s *jen.Statement) {
			if a.method {
				s.Call()
			}
		})

	default:
		fieldName = a.fieldNameN
		return jen.Id(argsWithIndex(a.fieldPrefix, a.count-1) + fieldName)
	}
}
