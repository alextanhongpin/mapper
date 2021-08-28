// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.
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

func (m *MapperImpl) AtoB(a0A A) B {
	a1A := m.mapMainAToMainB(a0A)
	return a1A
}

func (m *MapperImpl) ExternalAtoB(a0A []examples.A) []examples.B {
	a1A := make([]examples.B, len(a0A))
	for i, each := range a0A {
		a1A[i] = m.mapExamplesAToExamplesB(each)
	}
	return a1A
}

func (m *MapperImpl) SliceAtoB(a0A []A) []B {
	a1A := make([]B, len(a0A))
	for i, each := range a0A {
		a1A[i] = m.mapMainAToMainB(each)
	}
	return a1A
}

func (m *MapperImpl) Variadic(a0A []A) []B {
	a1A := make([]B, len(a0A))
	for i, each := range a0A {
		a1A[i] = m.mapMainAToMainB(each)
	}
	return a1A
}

func (m *MapperImpl) VariadicError(a0A []A) ([]B, error) {
	a1A := make([]B, len(a0A))
	for i, each := range a0A {
		a1A[i] = m.mapMainAToMainB(each)
	}
	return a1A, nil
}
