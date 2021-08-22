// Code gend by mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) B {
	return B{
		Bool:  a0.Bool,
		ID:    a0.ID,
		Map:   a0.Map,
		Slice: a0.Slice,
		Str:   a0.Str,
	}
}

func (m *MapperImpl) mapExamplesAToExamplesB(a0 examples.A) examples.B {
	return examples.B{
		Bool:  a0.Bool,
		ID:    a0.ID,
		Map:   a0.Map,
		Slice: a0.Slice,
		Str:   a0.Str,
	}
}

func (m *MapperImpl) SliceAtoB(a0 []A) []B {
	res := make([]B, len(a0))
	for i, each := range a0 {
		res[i] = m.mapMainAToMainB(each)
	}
	return res
}

func (m *MapperImpl) Variadic(a0 ...A) []B {
	res := make([]B, len(a0))
	for i, each := range a0 {
		res[i] = m.mapMainAToMainB(each)
	}
	return res
}

func (m *MapperImpl) VariadicError(a0 ...A) ([]B, error) {
	res := make([]B, len(a0))
	for i, each := range a0 {
		res[i] = m.mapMainAToMainB(each)
	}
	return res, nil
}

func (m *MapperImpl) AtoB(a0 A) B {
	return m.mapMainAToMainB(a0)
}

func (m *MapperImpl) ExternalAtoB(a0 []examples.A) []examples.B {
	res := make([]examples.B, len(a0))
	for i, each := range a0 {
		res[i] = m.mapExamplesAToExamplesB(each)
	}
	return res
}
