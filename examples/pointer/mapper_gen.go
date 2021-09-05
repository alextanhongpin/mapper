// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.
package main

import examples "github.com/alextanhongpin/mapper/examples"

var _ Mapper = (*MapperImpl)(nil)

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) B {
	a0NonPtrToPointer := m.mapMainCToMainD(a0.NonPtrToPointer)
	a1NonPtrToPointer := &a0NonPtrToPointer
	var a0Ptr *D
	if a0.Ptr != nil {
		tmp := m.mapMainCToMainD(*a0.Ptr)
		a0Ptr = &tmp
	}
	return B{
		Bool:            a0.Bool,
		ID:              a0.ID,
		Map:             a0.Map,
		NonPtrToPointer: a1NonPtrToPointer,
		Ptr:             a0Ptr,
		Slice:           a0.Slice,
		Str:             a0.Str,
	}
}

func (m *MapperImpl) mapMainCToMainD(c0 C) D {
	return D{
		Age:  &c0.Age,
		Name: c0.Name,
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

func (m *MapperImpl) AtoB(a0 A) *B {
	a1 := m.mapMainAToMainB(a0)
	a2 := &a1
	return a2
}

func (m *MapperImpl) CPointerToDPointer(c0 *C) *D {
	var c1 *D
	if c0 != nil {
		tmp := m.mapMainCToMainD(*c0)
		c1 = &tmp
	}
	return c1
}

func (m *MapperImpl) CtoD(c0 C) D {
	c1 := m.mapMainCToMainD(c0)
	return c1
}

func (m *MapperImpl) CtoDPointer(c0 C) *D {
	c1 := m.mapMainCToMainD(c0)
	c2 := &c1
	return c2
}

func (m *MapperImpl) ExternalAtoB(a0 examples.A) *examples.B {
	a1 := m.mapExamplesAToExamplesB(a0)
	a2 := &a1
	return a2
}
