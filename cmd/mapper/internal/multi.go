package internal

import "github.com/dave/jennifer/jen"

type Multi []*jen.Statement

func NewMulti(code ...*jen.Statement) *Multi {
	m := &Multi{}
	m.Add(code...)
	return m
}

func (m *Multi) Add(code ...*jen.Statement) *Multi {
	*m = append(*m, code...)
	return m
}

func (m *Multi) Statement() *jen.Statement {
	s := jen.Null()
	for i, n := range *m {
		if i != len(*m)-1 {
			s.Add(n).Line()
		} else {
			s.Add(n)
		}
	}
	return s
}
