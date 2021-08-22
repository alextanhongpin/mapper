// Code gend by mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) B {
	return B{
		Bool:  a0.Bool(),
		ID:    a0.ID(),
		Map:   a0.Map(),
		Slice: a0.Slice(),
		Str:   a0.Str(),
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

func (m *MapperImpl) AtoB(a0 A) B {
	return m.mapMainAToMainB(a0)
}

func (m *MapperImpl) ExternalAtoB(a0 examples.A) examples.B {
	return m.mapExamplesAToExamplesB(a0)
}
