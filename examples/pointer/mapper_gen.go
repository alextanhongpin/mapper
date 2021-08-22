// Code gend by mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) B {
	var a0Ptr *D
	if a0.Ptr != nil {
		res := m.mapMainCToMainD(*a0.Ptr)
		a0Ptr = &res
	}
	return B{
		Bool:  a0.Bool,
		ID:    a0.ID,
		Map:   a0.Map,
		Ptr:   a0Ptr,
		Slice: a0.Slice,
		Str:   a0.Str,
	}
}

func (m *MapperImpl) mapMainCToMainD(c0 C) D {
	return D{Name: c0.Name}
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

func (m *MapperImpl) AtoB(a0 A) *B {
	res := m.mapMainAToMainB(a0)
	return &res
}

func (m *MapperImpl) CtoD(c0 C) D {
	return m.mapMainCToMainD(c0)
}

func (m *MapperImpl) CtoDPointer(c0 C) *D {
	res := m.mapMainCToMainD(c0)
	return &res
}

func (m *MapperImpl) ExternalAtoB(a0 examples.A) *examples.B {
	res := m.mapExamplesAToExamplesB(a0)
	return &res
}
