// Code generated by github.com/alextanhongpin/mapper, DO NOT EDIT.
package main

var _ Mapper = (*MapperImpl)(nil)

type MapperImpl struct{}

func NewMapperImpl() *MapperImpl {
	return &MapperImpl{}
}

func (m *MapperImpl) mapMainAToMainB(g0 A) B {
	g0CustomStatus := g0.CustomStatus()
	return B{Status: g0CustomStatus}
}

func (m *MapperImpl) AtoB(g0 A) B {
	g1 := m.mapMainAToMainB(g0)
	return g1
}
