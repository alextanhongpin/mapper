// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.
package main

var _ Mapper = (*MapperImpl)(nil)

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(a0 A) B {
	a0Status := a0.CustomStatus()
	return B{Status: a0Status}
}

func (m *MapperImpl) AtoB(a0 A) B {
	a1 := m.mapMainAToMainB(a0)
	return a1
}
